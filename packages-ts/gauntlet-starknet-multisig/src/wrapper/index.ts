import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import {
  CommandCtor,
  Dependencies,
  ExecuteCommandInstance,
  ExecutionContext,
  Input,
  IStarknetProvider,
  IStarknetWallet,
} from '@chainlink/gauntlet-starknet'
import { TransactionResponse } from '@chainlink/gauntlet-starknet/dist/transaction'
import { Call, CompiledContract, Contract } from 'starknet'
import { getSelectorFromName } from 'starknet/dist/utils/hash'
import { toBN, toHex } from 'starknet/dist/utils/number'
import { contractLoader } from '../lib/contracts'
import { Action, State } from './types'

type UserInput = {
  proposalId: number
}

type ContractInput = {}

type UnregisteredCommand<UI, CI> = (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>

type ProposalAction = (message: Call, proposalId?: number) => Call

export const wrapCommand = <UI, CI>(
  registeredCommand: CommandCtor<ExecuteCommandInstance<UI, CI>>,
): UnregisteredCommand<UserInput, ContractInput> => (
  deps: Dependencies,
): CommandCtor<ExecuteCommandInstance<UserInput, ContractInput>> => {
  const id = `${registeredCommand.id}:multisig`

  const msigCommand: CommandCtor<ExecuteCommandInstance<UserInput, ContractInput>> = class MsigCommand
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
    state: State
    multisigContract: Contract

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
      c.wallet = deps.makeWallet(env.pk, env.account)
      c.contractAddress = args[0]
      c.account = env.account
      c.multisigAddress = env.multisig
      c.contract = contractLoader()
      c.multisigContract = new Contract(c.contract.abi, c.multisigAddress, c.provider.provider)

      c.executionContext = {
        provider: c.provider,
        wallet: c.wallet,
        id,
        contract: c.contractAddress,
        flags: flags,
      }

      c.input = {
        user: flags.input || { proposalId: Number(flags.proposalId || flags.multisigProposal) },
        contract: {},
      }

      c.command = await registeredCommand.create(flags, [c.contractAddress])

      c.state = await c.fetchMultisigState(c.multisigAddress, c.input.user.proposalId)

      return c
    }

    fetchMultisigState = async (address: string, proposalId?: number): Promise<State> => {
      const [owners, threshold] = await Promise.all(
        ['get_owners', 'get_confirmations_required'].map((func) => {
          return this.multisigContract[func]()
        }),
      )
      const multisig = {
        address,
        threshold: toBN(threshold.confirmations_required).toNumber(),
        owners: owners.owners.map((o) => toHex(o)),
      }

      if (isNaN(proposalId)) return { multisig }
      const proposal = await this.multisigContract.get_transaction(proposalId)
      return {
        multisig,
        proposal: {
          id: proposalId,
          approvers: proposal.tx.num_confirmations,
          data: proposal.tx_calldata,
          nextAction:
            toBN(proposal.tx.executed).toNumber() !== 0
              ? Action.NONE
              : proposal.tx.num_confirmations >= multisig.threshold
              ? Action.EXECUTE
              : Action.APPROVE,
        },
      }
    }

    makeProposeMessage: ProposalAction = (message) => {
      const invocation = this.multisigContract.populate('submit_transaction', [
        this.contractAddress,
        toBN(getSelectorFromName(message.entrypoint)),
        message.calldata,
      ])
      return invocation
    }

    makeAcceptMessage: ProposalAction = (_, proposalId) => {
      const invocation = this.multisigContract.populate('confirm_transaction', [proposalId])
      return invocation
    }

    makeExecuteMessage: ProposalAction = (_, proposalId) => {
      const invocation = this.multisigContract.populate('execute_transaction', [proposalId])
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
      if (!this.state.proposal) return [this.makeProposeMessage(message[0])]

      return [operations[this.state.proposal.nextAction](message[0], this.state.proposal.id)]
    }

    beforeExecute = async () => {
      console.log('Prompt what step we are and we are doing')
      await deps.prompt('Continue?')
    }

    afterExecute = async (result) => {
      console.log('Next steps')
      return {}
    }

    execute = async () => {
      deps.logger.info(`Multisig State:
        - Address: ${this.state.multisig.address}
        - Owners: ${this.state.multisig.owners}
        - Threshold: ${this.state.multisig.threshold}
      `)
      if (this.state.proposal) {
        deps.logger.info(`Proposal State:
        - ID: ${this.state.proposal.id}
        - Appovals: ${this.state.proposal.approvers}
        - Next action: ${this.state.proposal.nextAction}
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
      const data = await this.afterExecute(result)

      return !!data ? { ...result, data: { ...data } } : result
    }
  }

  return msigCommand
}
