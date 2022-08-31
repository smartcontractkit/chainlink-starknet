import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import { isDeepEqual } from '@chainlink/gauntlet-core/dist/utils/assertions'
import {
  CommandCtor,
  Dependencies,
  ExecuteCommandInstance,
  ExecutionContext,
  Input,
  IStarknetProvider,
  IStarknetWallet,
} from '@chainlink/starknet-gauntlet'
import { TransactionResponse } from '@chainlink/starknet-gauntlet/dist/transaction'
import {
  Call,
  CompiledContract,
  Contract,
  addAddressPadding,
  number,
  encode,
  validateAndParseAddress,
} from 'starknet'
import { addHexPrefix } from 'starknet/dist/utils/encode'
import { getSelectorFromName } from 'starknet/dist/utils/hash'
import { toBN, toHex } from 'starknet/dist/utils/number'
import { contractLoader } from '../lib/contracts'
import { Action, State } from './types'

type UserInput = {
  proposalId: number
}

type ContractInput = {}

type UnregisteredCommand<UI, CI> = (
  deps: Dependencies,
) => CommandCtor<ExecuteCommandInstance<UI, CI>>

type ProposalAction = (message: Call, proposalId?: number) => Call

export const wrapCommand = <UI, CI>(
  registeredCommand: CommandCtor<ExecuteCommandInstance<UI, CI>>,
): UnregisteredCommand<UserInput, ContractInput> => (
  deps: Dependencies,
): CommandCtor<ExecuteCommandInstance<UserInput, ContractInput>> => {
  const id = `${registeredCommand.id}:multisig`

  const msigCommand: CommandCtor<ExecuteCommandInstance<
    UserInput,
    ContractInput
  >> = class MsigCommand
    extends WriteCommand<TransactionResponse>
    implements ExecuteCommandInstance<UserInput, ContractInput> {
    wallet: IStarknetWallet
    provider: IStarknetProvider
    contractAddress: string
    account: string
    executionContext: ExecutionContext
    contract: CompiledContract

    input: Input<UserInput, ContractInput>

    command: ExecuteCommandInstance<UI, CI>
    multisigAddress: string
    initialState: State

    static id = id
    static category = registeredCommand.category
    static examples = registeredCommand.examples

    constructor(flags, args) {
      super(flags, args)
    }

    static create = async (flags, args) => {
      const c = new MsigCommand(flags, args)

      const env = deps.makeEnv(flags)

      c.provider = deps.makeProvider(env.providerUrl)
      c.wallet = await deps.makeWallet(env)
      c.contractAddress = args[0]
      c.account = env.account
      c.multisigAddress = env.multisig
      c.contract = contractLoader()

      c.executionContext = {
        provider: c.provider,
        wallet: c.wallet,
        id,
        contractAddress: c.contractAddress,
        flags: flags,
        contract: new Contract(c.contract.abi, c.multisigAddress, c.provider.provider),
      }

      c.input = {
        user: flags.input || {
          proposalId: Number(flags.proposalId || flags.multisigProposal),
        },
        contract: {},
      }

      c.command = await registeredCommand.create(flags, [c.contractAddress])

      c.initialState = await c.fetchMultisigState(c.multisigAddress, c.input.user.proposalId)

      return c
    }

    fetchMultisigState = async (address: string, proposalId?: number): Promise<State> => {
      const [signers, threshold] = await Promise.all(
        ['get_signers', 'get_threshold'].map((func) => {
          return this.executionContext.contract[func]()
        }),
      )
      const multisig = {
        address,
        threshold: toBN(threshold.threshold).toNumber(),
        signers: signers.signers.map((o) => toHex(o)),
      }

      if (isNaN(proposalId)) return { multisig }
      const proposal = await this.executionContext.contract.get_transaction(proposalId)
      return {
        multisig,
        proposal: {
          id: proposalId,
          confirmations: proposal.tx.confirmations,
          data: {
            contractAddress: addAddressPadding(proposal.tx.to),
            entrypoint: addHexPrefix(toBN(proposal.tx.function_selector).toString('hex')),
            calldata: proposal.tx_calldata.map((v) => toBN(v).toString()),
          },
          nextAction:
            toBN(proposal.tx.executed).toNumber() !== 0
              ? Action.NONE
              : proposal.tx.confirmations >= multisig.threshold
              ? Action.EXECUTE
              : Action.APPROVE,
        },
      }
    }

    isSameProposal = (local: Call, onchain: Call): boolean => {
      const localAddress = validateAndParseAddress(local.contractAddress)
      const onchainAddress = validateAndParseAddress(onchain.contractAddress)
      if (localAddress !== onchainAddress) return false
      if (getSelectorFromName(local.entrypoint) !== onchain.entrypoint) return false
      if (!isDeepEqual(local.calldata, onchain.calldata)) return false
      return true
    }

    makeProposeMessage: ProposalAction = (message, proposalId) => {
      const invocation = this.executionContext.contract.populate('submit_transaction', [
        this.contractAddress,
        toBN(getSelectorFromName(message.entrypoint)),
        message.calldata,
        proposalId,
      ])
      return invocation
    }

    makeAcceptMessage: ProposalAction = (_, proposalId) => {
      const invocation = this.executionContext.contract.populate('confirm_transaction', [
        proposalId,
      ])
      return invocation
    }

    makeExecuteMessage: ProposalAction = (_, proposalId) => {
      const invocation = this.executionContext.contract.populate('execute_transaction', [
        proposalId,
      ])
      return invocation
    }

    makeMessage = async () => {
      const operations = {
        [Action.APPROVE]: this.makeAcceptMessage,
        [Action.EXECUTE]: this.makeExecuteMessage,
        [Action.NONE]: () => {
          throw new Error('No action needed')
        },
      }
      const message = await this.command.makeMessage()
      if (!this.initialState.proposal) {
        const nonce = await this.executionContext.contract.get_transactions_len()
        return [this.makeProposeMessage(message[0], nonce)]
      }

      if (!this.isSameProposal(message[0], this.initialState.proposal.data)) {
        throw new Error('The transaction generated is different from the proposal provided')
      }

      deps.logger.success('Generated proposal matches with the one provided')

      return [
        operations[this.initialState.proposal.nextAction](
          message[0],
          this.initialState.proposal.id,
        ),
      ]
    }

    beforeExecute = async () => {
      if (!this.initialState.proposal) deps.logger.info('About to CREATE a new multisig proposal')
      else
        deps.logger.info(
          `About to ${
            this.initialState.proposal.nextAction === Action.APPROVE ? 'APPROVE' : 'EXECUTE'
          } the multisig proposal with ID: ${this.initialState.proposal.id}`,
        )
      await deps.prompt('Continue?')
    }

    afterExecute = async (result: Result<TransactionResponse>, proposalId?: number) => {
      if (proposalId === undefined) deps.logger.warn('Proposal ID not found')
      deps.logger.loading('Fetching latest multisig and proposal state...')
      const state = await this.fetchMultisigState(this.multisigAddress, proposalId)
      if (!state.proposal) {
        deps.logger.error(`Multisig proposal ${proposalId} not found`)
        return
      }
      const approvalsLeft = state.multisig.threshold - state.proposal.confirmations
      const messages = {
        [Action.EXECUTE]: `The multisig proposal reached the threshold and can be executed. Run the same command with the flag --multisigProposal=${state.proposal.id}`,
        [Action.APPROVE]: `The multisig proposal needs ${approvalsLeft} more approvals. Run the same command with the flag --multisigProposal=${state.proposal.id}`,
        [Action.NONE]: `The multisig proposal has been executed. No more actions needed`,
      }
      deps.logger.line()
      deps.logger.info(`${messages[state.proposal.nextAction]}`)
      deps.logger.line()
      return { proposalId }
    }

    execute = async () => {
      deps.logger.info(`Multisig State:
        - Address: ${this.initialState.multisig.address}
        - Signers: ${this.initialState.multisig.signers}
        - Threshold: ${this.initialState.multisig.threshold}
      `)
      if (this.initialState.proposal) {
        deps.logger.info(`Proposal State:
        - ID: ${this.initialState.proposal.id}
        - Approvals: ${this.initialState.proposal.confirmations}
        - Next action: ${this.initialState.proposal.nextAction}
      `)
      }
      const message = await this.makeMessage()

      // Underlying logger could have different style and probably a disabled prompt
      await this.command.beforeExecute()
      await this.beforeExecute()

      deps.logger.loading(`Signing and sending transaction...`)
      const tx = await this.provider.signAndSend(this.account, this.wallet, message)
      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()

      if (!response.success) {
        deps.logger.error(`Tx was not successful: ${tx.errorMessage}`)
      } else {
        deps.logger.success(`Tx executed at ${tx.hash}`)
      }

      let result = {
        responses: [
          {
            tx,
            contract: tx.address,
          },
        ],
      }

      let proposalId = this.input.user.proposalId
      if (!this.initialState.proposal) {
        const txInfo = await this.provider.provider.getTransactionReceipt(tx.hash)
        proposalId = Number(number.hexToDecimalString((txInfo.events[0] as any).data[1]))
      }

      const data = await this.afterExecute(result, proposalId)

      return !!data ? { ...result, data: { ...data } } : result
    }
  }

  return msigCommand
}
