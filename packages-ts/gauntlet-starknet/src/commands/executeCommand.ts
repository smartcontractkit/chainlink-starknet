import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import { RawCalldata, CompiledContract, Contract } from 'starknet'
import { Dependencies } from '../dependencies'
import { IStarknetProvider, wrapResponse } from '../provider'
import { TransactionResponse } from '../transaction'
import { IStarknetWallet } from '../wallet'
import { Validation } from './command'

export interface ExecutionContext {
  id: string
  contract: string
  wallet: IStarknetWallet
  provider: IStarknetProvider
  flags: any
}

export interface Input<UI, CI> {
  user: UI
  contract: CI
}

export type BeforeExecute<UI, CI> = (
  context: ExecutionContext,
  input: Input<UI, CI>,
  deps: Pick<Dependencies, 'logger' | 'prompt'>,
) => (signer: string) => Promise<void>

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
    beforeExecute: BeforeExecute<UI, CI>
    afterExecute: AfterExecute<UI, CI>
  }
  makeUserInput: (flags, args) => Promise<UI>
  makeContractInput: (userInput: UI) => Promise<CI>
  validations: Validation<UI>[]
  contract: CompiledContract
}

export interface ExecuteCommandInstance {
  makeMessage: (signer: string) => RawCalldata
  execute: () => Promise<Result<TransactionResponse>>
  simulate?: () => boolean
}

const makeCommandId = (category: string, fn: string, suffixes?: string[]): string => {
  const base = `${category}:${fn}`
  return suffixes?.length > 0 ? `${base}:${suffixes.join(':')}` : base
}

export const makeExecuteCommand = <UI, CI>(config: ExecuteCommandConfig<UI, CI>) => (deps: Dependencies) => {
  return class ExecuteCommand extends WriteCommand<TransactionResponse> implements ExecuteCommandInstance {
    wallet: IStarknetWallet
    provider: IStarknetProvider
    contractAddress: string
    executionContext: ExecutionContext

    input: Input<UI, CI>

    beforeExecute: (signer: string) => Promise<void>
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

      c.executionContext = {
        provider: c.provider,
        wallet: c.wallet,
        id: makeCommandId(config.ux.category, config.ux.function, config.ux.suffixes),
        contract: c.contractAddress,
        flags: flags,
      }

      const input = await c.buildCommandInput(flags, args)

      c.input = input

      c.beforeExecute = config.hooks.beforeExecute
        ? config.hooks.beforeExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.defaultBeforeExecute(c.executionContext, c.input)

      c.afterExecute = config.hooks.afterExecute
        ? config.hooks.afterExecute(c.executionContext, c.input, { logger: deps.logger, prompt: deps.prompt })
        : c.afterExecute

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
      deps.logger.log('Input Params:', input.contract)
    }

    defaultAfterExecute = () => async (response: Result<TransactionResponse>): Promise<any> => {
      deps.logger.success(`Execution finished at transaction: ${response.responses[0].tx.hash}`)
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

    makeMessage = (signer) => {
      return Array.from(Buffer.from(''))
    }

    deployContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Deploying contract ${config.ux.category}`)
      await deps.prompt('Continue?')

      const tx = await this.provider.deployContract(config.contract, false)
      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      await tx.wait()
      deps.logger.success(`Contract deployed on ${tx.hash} with address ${tx.address}`)
      return tx
    }

    executeFn = async (): Promise<TransactionResponse> => {
      const contract = new Contract(config.contract.abi, this.contractAddress, this.provider.provider)
      deps.logger.info(
        `Executing function ${config.ux.function} of contract ${config.ux.category} at ${this.contractAddress}`,
      )
      deps.logger.log('Contract Input:', this.input.contract)
      await deps.prompt('Continue?')

      const tx = await contract[config.ux.function](...(this.input.contract as any))
      const response = wrapResponse(this.provider, tx, this.contractAddress)
      deps.logger.loading(`Waiting for tx confirmation at ${response.hash}...`)
      await response.wait()
      deps.logger.success(`Function executed on ${response.hash}`)
      return response
    }

    execute = async () => {
      let tx: TransactionResponse

      const wallet = await this.wallet.wallet.getPubKey()
      await this.beforeExecute(wallet.toString())

      if (config.ux.function === 'deploy') {
        tx = await this.deployContract()
      } else {
        tx = await this.executeFn()
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
}
