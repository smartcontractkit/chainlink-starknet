import { Result, WriteCommand, BaseConfig } from '@chainlink/gauntlet-core'
import {
  CompiledContract,
  CompiledSierraCasm,
  Contract,
  Call,
  hash,
  DeclareContractResponse,
} from 'starknet'
import { CommandCtor } from '.'
import { Dependencies } from '../../dependencies'
import { IStarknetProvider } from '../../provider'
import { getRDD } from '../../rdd'
import { TransactionResponse } from '../../transaction'
import { IStarknetWallet } from '../../wallet'
import { makeCommandId, Validation, Input } from './command'

export interface ExecutionContext {
  category: string
  action: string
  id: string
  contractAddress: string
  wallet: IStarknetWallet
  provider: IStarknetProvider
  flags: any
  contract: Contract
  rdd?: any
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

export interface LoadContractResult {
  contract: CompiledContract
  // for cairo 1.0, `contract` is a sierra artifact and the casm artifact needs to be provided.
  casm?: CompiledSierraCasm
}

export interface ExecuteCommandConfig<UI, CI> extends BaseConfig<UI> {
  hooks?: {
    beforeExecute?: BeforeExecute<UI, CI>
    afterExecute?: AfterExecute<UI, CI>
  }
  internalFunction?: string
  makeContractInput: (userInput: UI, context: ExecutionContext) => Promise<CI>
  loadContract: () => LoadContractResult
}

export interface ExecuteCommandInstance<UI, CI> {
  wallet: IStarknetWallet
  provider: IStarknetProvider
  contractAddress: string
  account: string
  executionContext: ExecutionContext
  contract: CompiledContract
  compiledContractHash?: string

  input: Input<UI, CI>

  makeMessage: () => Promise<Call[]>
  execute: () => Promise<Result<TransactionResponse>>
  simulate?: () => boolean

  beforeExecute: () => Promise<void>
  afterExecute: (response: Result<TransactionResponse>) => Promise<any>
}

export const makeExecuteCommand = <UI, CI>(config: ExecuteCommandConfig<UI, CI>) => (
  deps: Dependencies,
) => {
  const command: CommandCtor<ExecuteCommandInstance<UI, CI>> = class ExecuteCommand
    extends WriteCommand<TransactionResponse>
    implements ExecuteCommandInstance<UI, CI> {
    wallet: IStarknetWallet
    provider: IStarknetProvider
    contractAddress: string
    account: string
    executionContext: ExecutionContext
    contract: CompiledContract
    compiledContractHash?: string

    input: Input<UI, CI>

    beforeExecute: () => Promise<void>
    afterExecute: (response: Result<TransactionResponse>) => Promise<any>

    static id = makeCommandId(config.category, config.action, config.suffixes)
    static category = config.category
    static examples = config.ux.examples

    static create = async (flags, args) => {
      const c = new ExecuteCommand(flags, args)

      const env = await deps.makeEnv(flags)

      c.wallet = await deps.makeWallet(env)
      c.provider = deps.makeProvider(env.providerUrl, c.wallet)
      c.contractAddress = args[0]
      c.account = env.account

      const loadResult = config.loadContract()
      c.contract = loadResult.contract
      if (loadResult.casm) {
        c.compiledContractHash = hash.computeCompiledClassHash(loadResult.casm)
      }

      c.executionContext = {
        category: config.category,
        action: config.action,
        provider: c.provider,
        wallet: c.wallet,
        id: makeCommandId(config.category, config.action, config.suffixes),
        contractAddress: c.contractAddress,
        flags: flags,
        contract: new Contract(c.contract.abi, c.contractAddress ?? '', c.provider.provider),
      }

      const rdd = flags.rdd || process.env.RDD
      if (rdd) {
        deps.logger.info(`Using RDD from ${rdd}`)
        c.executionContext.rdd = getRDD(rdd)
      }

      c.input = await c.buildCommandInput(flags, args, env)

      c.beforeExecute = config.hooks?.beforeExecute
        ? config.hooks.beforeExecute(c.executionContext, c.input, {
            logger: deps.logger,
            prompt: deps.prompt,
          })
        : c.defaultBeforeExecute(c.executionContext, c.input)

      c.afterExecute = config.hooks?.afterExecute
        ? config.hooks.afterExecute(c.executionContext, c.input, {
            logger: deps.logger,
            prompt: deps.prompt,
          })
        : c.defaultAfterExecute()

      return c
    }

    runValidations = async (validations: Validation<UI, ExecutionContext>[], input: UI) => {
      const result = await Promise.all(
        validations.map((validation) => validation(input, this.executionContext)),
      )
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
      const makeInvocation = (contractAddress: string) => {
        const contract = new Contract(this.contract.abi, contractAddress, this.provider.provider)
        return contract.populate(
          config.internalFunction || config.action,
          this.input.contract as any,
        )
      }

      // If a flag called `batch` is present and it's value is `true`,
      // then the following assumptions will be made here:
      //
      //  - All `args` are valid Starknet contract addresses
      //  - All contract addresses reference contracts with the same ABI
      //  - All contract invocations will be populated with the same inputs
      //
      const batch = this.flags.batch
      if (typeof batch === 'boolean' && batch) {
        return this.args.map(makeInvocation)
      }

      return [makeInvocation(this.contractAddress)]
    }

    deployContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Deploying contract ${config.category}`)
      await deps.prompt('Continue?')
      deps.logger.loading(`Sending transaction...`)

      // if "--classHash" is not included, declare before deploying
      const classHash: string | undefined = this.input?.user?.['classHash']

      let tx: TransactionResponse

      if (classHash === undefined) {
        tx = await this.provider.declareAndDeployContract(
          this.contract,
          this.compiledContractHash,
          this.input.contract,
          false,
          this.input?.user?.['salt'],
        )
      } else {
        tx = await this.provider.deployContract(
          classHash,
          this.input.contract,
          false,
          this.input?.user?.['salt'],
        )
      }

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
      deps.logger.info(
        `If using RDD, change the RDD ID with the new contract address: ${tx.address}`,
      )
      return tx
    }

    declareContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Declaring contract ${config.category}`)
      await deps.prompt('Continue?')
      deps.logger.loading(`Sending transaction...`)

      const tx = await this.provider.declareContract(
        this.contract,
        this.compiledContractHash,
        false,
      )

      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()

      if (!response.success) {
        deps.logger.error(`Contract was not declared: ${tx.errorMessage}`)
        return tx
      }

      deps.logger.success(`Contract declared at ${this.compiledContractHash}`)
      return tx
    }

    deployAccountContract = async (): Promise<TransactionResponse> => {
      deps.logger.info(`Deploying account contract ${config.category}`)
      await deps.prompt('Continue?')
      deps.logger.loading(`Sending transaction...`)

      // classHash has to be provided, we can't declare on new accounts.
      const classHash: string = this.input?.user?.['classHash']
      const salt = this.input?.user?.['salt']
      const contractInput = this.input.contract as any

      const newAccountAddress = hash.calculateContractAddressFromHash(
        salt,
        classHash,
        contractInput,
        0,
      )
      deps.logger.info(
        `Add funds to pay for deploy fees to the account address: ${newAccountAddress}`,
      )
      await deps.prompt('Funded?')

      const tx: TransactionResponse = await this.provider.deployAccountContract(
        classHash,
        this.input.contract,
        false,
        salt,
      )

      if (tx.hash === undefined) {
        deps.logger.error(`No tx hash found: \n${JSON.stringify(tx, null, 2)}`)
        return tx
      }

      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()
      if (!response.success) {
        deps.logger.error(`Contract was not deployed: ${tx.errorMessage}`)
        return tx
      }
      deps.logger.success(`Contract deployed on ${tx.hash} with address ${tx.address}`)
      deps.logger.info(
        `If using RDD, change the RDD ID with the new contract address: ${tx.address}`,
      )
      return tx
    }

    executeWithSigner = async (): Promise<TransactionResponse> => {
      const pubkey = await this.wallet.getPublicKey()
      deps.logger.info(`Using wallet: ${pubkey}`)
      const messages = await this.makeMessage()
      await deps.prompt(`Continue?`)
      deps.logger.loading(`Signing and sending transaction...`)
      const tx = await this.provider.signAndSend(messages)
      deps.logger.loading(`Waiting for tx confirmation at ${tx.hash}...`)
      const response = await tx.wait()
      if (!response.success) {
        deps.logger.error(`Tx was not successful: ${tx.errorMessage}`)
        return tx
      }
      deps.logger.success(`Tx executed at ${tx.hash}`)
      return tx
    }

    execute = async () => {
      let tx: TransactionResponse

      await this.beforeExecute()

      if (config.action === 'deploy') {
        tx = await this.deployContract()
      } else if (config.action === 'declare') {
        tx = await this.declareContract()
      } else if (config.action === 'deploy-account') {
        tx = await this.deployAccountContract()
      } else {
        tx = await this.executeWithSigner()
      }

      const result = {
        responses: [
          {
            tx,
            contract: tx.address,
          },
        ],
      }
      const data = await this.afterExecute(result)

      return data ? { ...result, data: { ...data } } : result
    }
  }

  return command
}
