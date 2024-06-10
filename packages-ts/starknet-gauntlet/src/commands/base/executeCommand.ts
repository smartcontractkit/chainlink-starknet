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
  batchInput?: Array<[string, Input<UI, CI>]>

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
    batchInput?: Array<[string, Input<UI, CI>]>

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
      c.account = env.account

      // NOTE: If more than one argument is passed in, then the following assumptions will be made:
      //  - All `args` are valid Starknet contract addresses
      //  - All contract addresses reference contracts with the same ABI
      //
      c.contractAddress = args[0]

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

      const overrideExecutionContext = (contractAddress: string): ExecutionContext => {
        return {
          ...c.executionContext,
          contract: new Contract(c.contract.abi, contractAddress, c.provider.provider),
          contractAddress,
        }
      }

      const rdd = flags.rdd || process.env.RDD
      if (rdd) {
        deps.logger.info(`Using RDD from ${rdd}`)
        c.executionContext.rdd = getRDD(rdd)

        // If more than one argument / address is passed to the command, then we'll assume that we
        // want to perform a batch operation where we execute multiple calls from one transaction.
        // If the RDD flag is also specified, then we gain the flexibility to vary the inputs of each
        // call. Instead of forwarding all args to `buildCommandInput`, we can pass each arg to the
        // function individually:
        //
        //   makeBuildInput(..., [args[0]], ..., <context with contract object and contract address replaced with args[0]>)
        //   makeBuildInput(..., [args[1]], ..., <context with contract object and contract address replaced with args[1]>)
        //   makeBuildInput(..., [args[2]], ..., <context with contract object and contract address replaced with args[2]>)
        //   ...
        //
        // Under the hood, each call will access the RDD JSON config file at key `args[i]` which will
        // return a different set of config values. These values can be used to populate the transaction
        // calls.
        //
        if (args.length > 1) {
          // Maps each address to its corresponding input while preserving
          // the order in which the addresses were passed to the command
          c.batchInput = await Promise.all(
            args.map(async (addr: string) => {
              const execCtx = overrideExecutionContext(addr)
              const cmdInpt = await c.buildCommandInput(flags, [addr], env, execCtx)
              return [addr, cmdInpt] as const
            }),
          )
        }
      }

      c.input = await c.buildCommandInput(flags, args, env)

      // Map each address to its corresponding input using an array of pairs
      const addressToInputPairs = c.batchInput ?? ([[c.contractAddress, c.input]] as const)

      c.beforeExecute = async () => {
        const funcs = addressToInputPairs.map(([addr, inpt]) => {
          const ctx = overrideExecutionContext(addr)
          return config.hooks?.beforeExecute == null
            ? c.defaultBeforeExecute(ctx, inpt)
            : config.hooks.beforeExecute(ctx, inpt, {
              logger: deps.logger,
              prompt: deps.prompt,
            })
        })
        for (const func of funcs) {
          await func()
        }
      }

      c.afterExecute = async (result) => {
        const funcs = addressToInputPairs.map(([addr, inpt]) => {
          const ctx = overrideExecutionContext(addr)
          return config.hooks?.afterExecute == null
            ? c.defaultAfterExecute()
            : config.hooks.afterExecute(ctx, inpt, {
              logger: deps.logger,
              prompt: deps.prompt,
            })
        })

        // If there's only one function to call (bc only one address was provided to
        // the command), then we will NOT wrap the result in an array. Instead we'll
        // return the result as-is to maintain backwards compatibility.
        const firstFunc = funcs.at(0)
        if (firstFunc != null && funcs.length === 1) {
          return await firstFunc(result)
        }

        // If we need to execute multiple functions (bc multiple addresses were provided),
        // then we will call each of them in serial and in the order in which the input
        // addresses were provided. The outputs of each function are stored in an array.
        const outputs = []
        for (const func of funcs) {
          outputs.push(await func(result))
        }
        return outputs
      }

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

    buildCommandInput = async (flags, args, env, ctx?): Promise<Input<UI, CI>> => {
      const userInput = await config.makeUserInput(flags, args, env)

      // Validation
      if (config.validations.length > 0) {
        await this.runValidations(config.validations, userInput)
      }

      const contractInput = await config.makeContractInput(userInput, ctx ?? this.executionContext)

      return {
        user: userInput,
        contract: contractInput,
      }
    }

    simulate = () => true

    // TODO: This will be required for Multisig
    makeMessage = async (): Promise<Call[]> => {
      const makeInvocation = (addr: string, input: any) => {
        const contract = new Contract(this.contract.abi, addr, this.provider.provider)
        return contract.populate(config.internalFunction || config.action, input)
      }

      if (this.args.length === 0) {
        return []
      }

      if (this.args.length === 1) {
        return [makeInvocation(this.contractAddress, this.input.contract)]
      }

      // If we have more than one argument / address and the RDD flag is set, then we'll build
      // one transaction with multiple calls where each call has potentially different inputs.
      const inputs = this.batchInput
      if (inputs != null) {
        return inputs.map(([addr, inpt]) => makeInvocation(addr, inpt.contract))
      }

      // If more than one argument / address is passed to the command, but the RDD flag is NOT
      // specified, then we cannot vary the inputs of each call. Instead, all calls must have
      // the same arguments. However, they will still be bundled into one transaction.
      return this.args.map((addr) => makeInvocation(addr, this.input.contract))
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
