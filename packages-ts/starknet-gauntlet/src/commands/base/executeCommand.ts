import { Result, WriteCommand, BaseConfig } from '@chainlink/gauntlet-core'
import { CompiledContract, Contract, Call } from 'starknet'
import { CommandCtor } from '.'
import { Dependencies, Env } from '../../dependencies'
import { IStarknetProvider, wrapResponse } from '../../provider'
import { TransactionResponse } from '../../transaction'
import { IStarknetWallet } from '../../wallet'
import { makeCommandId, Validation, Input } from './command'

export interface ExecutionContext {
  id: string
  contractAddress: string
  wallet: IStarknetWallet
  provider: IStarknetProvider
  flags: any
  contract: Contract
}

export type BeforeExecute<UI, CI> = (
  context: ExecutionContext,
  input: Input<UI, CI>,
  deps: Pick<Dependencies, 'logger' | 'prompt'>,
) => () => Promise<void>

export type AfterExecute<UI, CI> = (
  context: ExecutionContext,
  input: Input<UI, CI>,
  deps: Pick<Dependencies, 'logger' | 'prompt'>,
) => (result: Result<TransactionResponse>) => Promise<any>

export interface ExecuteCommandConfig<UI, CI> extends BaseConfig<UI> {
  hooks?: {
    beforeExecute?: BeforeExecute<UI, CI>
    afterExecute?: AfterExecute<UI, CI>
  }
  internalFunction?: string
  makeContractInput: (userInput: UI, context: ExecutionContext) => Promise<CI>
  loadContract: () => CompiledContract
}

export interface ExecuteCommandInstance<UI, CI> {
  wallet: IStarknetWallet
  provider: IStarknetProvider
  contractAddress: string
  account: string
  executionContext: ExecutionContext
  contract: CompiledContract

  input: Input<UI, CI>

  makeMessage: () => Promise<Call[]>
  execute: () => Promise<Result<TransactionResponse>>
  simulate?: () => boolean

  beforeExecute: () => Promise<void>
  afterExecute: (response: Result<TransactionResponse>) => Promise<any>
}

export const makeExecuteCommand = <UI, CI>(config: ExecuteCommandConfig<UI, CI>) => (deps: Dependencies) => {
  const command: CommandCtor<ExecuteCommandInstance<UI, CI>> = class ExecuteCommand
    extends WriteCommand<TransactionResponse>
    implements ExecuteCommandInstance<UI, CI> {
    wallet: IStarknetWallet
    provider: IStarknetProvider
    contractAddress: string
    account: string
    executionContext: ExecutionContext
    contract: CompiledContract

    input: Input<UI, CI>

    beforeExecute: () => Promise<void>
    afterExecute: (response: Result<TransactionResponse>) => Promise<any>

    static id = makeCommandId(config.category, config.action, config.suffixes)
    static category = config.category
    static examples = config.ux.examples

    static create = async (flags, args) => {
      const c = new ExecuteCommand(flags, args)

      const env = deps.makeEnv(flags)

      c.provider = deps.makeProvider(env.providerUrl)
      c.wallet = deps.makeWallet(env.pk, env.account)
      c.contractAddress = args[0]
      c.account = env.account
      c.contract = config.loadContract()

      c.executionContext = {
        provider: c.provider,
        wallet: c.wallet,
        id: makeCommandId(config.category, config.action, config.suffixes),
        contractAddress: c.contractAddress,
        flags: flags,
        contract: new Contract(c.contract.abi, c.contractAddress, c.provider.provider),
      }

      c.input = await c.buildCommandInput(flags, args, env)

      c.beforeExecute = config.hooks?.beforeExecute
        ? config.hooks.beforeExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.defaultBeforeExecute(c.executionContext, c.input)

      c.afterExecute = config.hooks?.afterExecute
        ? config.hooks.afterExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.defaultAfterExecute()

      return c
    }

    runValidations = async (validations: Validation<UI, ExecutionContext>[], input: UI) => {
      const result = await Promise.all(validations.map((validation) => validation(input, this.executionContext)))
      return result
    }

    defaultBeforeExecute = <UserInput, ContractInput>(
      context: ExecutionContext,
      input: Input<UserInput, ContractInput>,
    ) => async () => {
      deps.logger.loading(`Executing ${context.id} from contract ${context.contractAddress}`)
      deps.logger.log('Contract Input Params:', input.contract)
    }

    defaultAfterExecute = () => async (response: Result<TransactionResponse>): Promise<any> => {
      deps.logger.info(`Execution finished at transaction: ${response.responses[0].tx.hash}`)
    }

    buildCommandInput = async (flags, args, env): Promise<Input<UI, CI>> => {
      const userInput = await config.makeUserInput(flags, args, env)

      // Validation
      if (config.validations.length > 0) {
        await this.runValidations(config.validations, userInput)
      }

      const contractInput = await config.makeContractInput(userInput, this.executionContext)

      return {
        user: userInput,
        contract: contractInput,
      }
    }

    simulate = () => true

    // TODO: This will be required for Multisig
    makeMessage = async (): Promise<Call[]> => {
      const contract = new Contract(this.contract.abi, this.contractAddress, this.provider.provider)
      const invocation = await contract.populate(config.internalFunction || config.action, this.input.contract as any)

      return [invocation]
    }

    deployContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Deploying contract ${config.category}`)
      await deps.prompt('Continue?')
      deps.logger.loading(`Sending transaction...`)

      const tx = await this.provider.deployContract(this.contract, this.input.contract, false, this.input.user['salt'])
      if (tx.hash === undefined) {
        deps.logger.error(`No tx hash found:\n${JSON.stringify(tx, null, 2)}`)
        return tx
      }

      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()
      if (!response.success) {
        deps.logger.error(`Contract was not deployed: ${tx.errorMessage}`)
        return tx
      }
      deps.logger.success(`Contract deployed on ${tx.hash} with address ${tx.address}`)
      return tx
    }

    executeWithSigner = async (): Promise<TransactionResponse> => {
      const pubkey = await this.wallet.getPublicKey()
      deps.logger.info(`Using wallet: ${pubkey}`)
      const messages = await this.makeMessage()
      await deps.prompt(`Continue?`)
      deps.logger.loading(`Signing and sending transaction...`)
      const tx = await this.provider.signAndSend(this.account, this.wallet, messages)
      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()
      if (!response.success) {
        deps.logger.error(`Tx was not successful: ${tx.errorMessage}`)
        return tx
      }
      deps.logger.success(`Tx executed at ${tx.hash}`)
      return tx
    }

    executeWithoutSigner = async (): Promise<TransactionResponse> => {
      const contract = new Contract(this.contract.abi, this.contractAddress, this.provider.provider)
      await deps.prompt(`Continue?`)
      deps.logger.loading(`Sending transaction...`)
      const tx = await contract[config.internalFunction || config.action](...(this.input.contract as any))
      const response = wrapResponse(this.provider, tx, this.contractAddress)
      deps.logger.loading(`Waiting for tx confirmation at ${response.hash}...`)
      await response.wait()
      return response
    }

    execute = async () => {
      let tx: TransactionResponse

      await this.beforeExecute()

      if (config.action === 'deploy') {
        tx = await this.deployContract()
      } else {
        if (this.flags.noWallet) {
          tx = await this.executeWithoutSigner()
        } else {
          tx = await this.executeWithSigner()
        }
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

  return command
}
