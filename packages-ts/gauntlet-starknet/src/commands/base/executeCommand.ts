import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import { CompiledContract, Contract, Call } from 'starknet'
import { CommandCtor } from '.'
import { Dependencies } from '../../dependencies'
import { IStarknetProvider, wrapResponse } from '../../provider'
import { TransactionResponse } from '../../transaction'
import { IStarknetWallet } from '../../wallet'
import { makeCommandId, Validation, Input } from './command'

export interface ExecutionContext {
  id: string
  contract: string
  wallet: IStarknetWallet
  provider: IStarknetProvider
  flags: any
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

export interface ExecuteCommandConfig<UI, CI> {
  ux: {
    category: string
    function: string
    suffixes?: string[]
    examples: string[]
  }
  hooks?: {
    beforeExecute?: BeforeExecute<UI, CI>
    afterExecute?: AfterExecute<UI, CI>
  }
  makeUserInput: (flags, args) => Promise<UI>
  makeContractInput: (userInput: UI) => Promise<CI>
  validations: Validation<UI>[]
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

    static id = makeCommandId(config.ux.category, config.ux.function, config.ux.suffixes)
    static category = config.ux.category
    static examples = config.ux.examples

    static create = async (flags, args) => {
      const c = new ExecuteCommand(flags, args)

      const env = deps.makeEnv(flags)

      c.provider = deps.makeProvider(env.providerUrl)
      c.wallet = deps.makeWallet(env.pk)
      c.contractAddress = args[0]
      c.account = env.account

      c.executionContext = {
        provider: c.provider,
        wallet: c.wallet,
        id: makeCommandId(config.ux.category, config.ux.function, config.ux.suffixes),
        contract: c.contractAddress,
        flags: flags,
      }

      c.input = await c.buildCommandInput(flags, args)
      c.contract = config.loadContract()

      c.beforeExecute = config.hooks?.beforeExecute
        ? config.hooks.beforeExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.defaultBeforeExecute(c.executionContext, c.input)

      c.afterExecute = config.hooks?.afterExecute
        ? config.hooks.afterExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.defaultAfterExecute()

      return c
    }

    runValidations = async (validations: Validation<UI>[], input: UI) => {
      const result = await Promise.all(validations.map((validation) => validation(input)))
      return result
    }

    defaultBeforeExecute = <UserInput, ContractInput>(
      context: ExecutionContext,
      input: Input<UserInput, ContractInput>,
    ) => async () => {
      deps.logger.loading(`Executing ${context.id} from contract ${context.contract}`)
      deps.logger.log('Contract Input Params:', input.contract)
      await deps.prompt('Continue?')
    }

    defaultAfterExecute = () => async (response: Result<TransactionResponse>): Promise<any> => {
      deps.logger.info(`Execution finished at transaction: ${response.responses[0].tx.hash}`)
    }

    buildCommandInput = async (flags, args): Promise<Input<UI, CI>> => {
      const userInput = await config.makeUserInput(flags, args)

      // Validation
      if (config.validations.length > 0) {
        await this.runValidations(config.validations, userInput)
      }

      const contractInput = await config.makeContractInput(userInput)

      return {
        user: userInput,
        contract: contractInput,
      }
    }

    simulate = () => true

    // TODO: This will be required for Multisig
    makeMessage = async (): Promise<Call[]> => {
      const contract = new Contract(this.contract.abi, this.contractAddress, this.provider.provider)
      const invocation = await contract.populate(config.ux.function, this.input.contract as any)

      return [invocation]
    }

    deployContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Deploying contract ${config.ux.category}`)
      await deps.prompt('Continue?')

      const tx = await this.provider.deployContract(this.contract, this.input.contract, false)
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
      const messages = await this.makeMessage()
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
      const tx = await contract[config.ux.function](...(this.input.contract as any))
      const response = wrapResponse(this.provider, tx, this.contractAddress)
      deps.logger.loading(`Waiting for tx confirmation at ${response.hash}...`)
      await response.wait()
      return response
    }

    execute = async () => {
      let tx: TransactionResponse

      const pubkey = await this.wallet.getPublicKey()
      deps.logger.info(`Using wallet: ${pubkey}`)

      await this.beforeExecute()

      if (config.ux.function === 'deploy') {
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
