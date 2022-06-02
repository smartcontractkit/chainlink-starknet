import { Result, WriteCommand } from '@chainlink/gauntlet-core'
import {
  CommandCtor,
  Dependencies,
  ExecuteCommandConfig,
  ExecuteCommandInstance,
  ExecutionContext,
  Input,
  IStarknetProvider,
  IStarknetWallet,
  makeExecuteCommand,
} from '@chainlink/gauntlet-starknet'
import { TransactionResponse } from '@chainlink/gauntlet-starknet/dist/transaction'
import { CompiledContract } from 'starknet'
import { registerExecuteCommand } from '..'
import { CATEGORIES } from '../lib/categories'

type UserInput = {}
type ContractInput = {}

type UnregisteredCommand<UI, CI> = (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>

const wrapCommand = <UI, CI>(command: UnregisteredCommand<UI, CI>): UnregisteredCommand<UserInput, ContractInput> => (
  deps: Dependencies,
): CommandCtor<ExecuteCommandInstance<UserInput, ContractInput>> => {
  const registeredCommand = registerExecuteCommand(command)

  const msigCommand: CommandCtor<ExecuteCommandInstance<UI, CI>> = class MsigCommand
    extends WriteCommand<TransactionResponse>
    implements ExecuteCommandInstance<UI, CI> {
    wallet: IStarknetWallet
    provider: IStarknetProvider
    contractAddress: string
    account: string
    executionContext: ExecutionContext
    contract: CompiledContract

    input: Input<UI, CI>

    static id = `${registeredCommand}:multisig`
    static category = registeredCommand.category
    static examples = registeredCommand.examples

    constructor(flags, args) {
      super(flags, args)
    }

    static create = async (flags, args) => {
      return new MsigCommand(flags, args)
    }

    makeMessage = async () => []
    execute = async () => {
      return {} as Result<TransactionResponse>
    }
  }

  return msigCommand
}
