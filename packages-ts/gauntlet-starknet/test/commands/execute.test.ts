import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  Dependencies,
  CommandCtor,
  ExecuteCommandInstance,
  makeProvider,
  makeWallet,
  ExecuteCommandConfig,
  makeExecuteCommand,
} from '../../src/index'
import { loadExampleContract, noopLogger, noopPrompt } from '../utils'

const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: logger,
    prompt: prompt,
    makeEnv: (flags) => {
      return {
        providerUrl: 'http://127.0.0.1:5000',
        pk: '',
        account: '',
      }
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

describe('Execute Command', () => {
  type UserInput = {
    a: string
    b: number
  }

  type ContractInput = [string, number]

  const makeUserInput = async (flags, args): Promise<UserInput> => {
    return {
      a: flags.a,
      b: Number(flags.b),
    }
  }

  const makeContractInput = async (userInput: UserInput): Promise<ContractInput> => {
    return [userInput.a, userInput.b]
  }

  const simpleCommandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
    ux: {
      category: 'example',
      function: 'action',
      examples: [],
    },
    makeUserInput,
    makeContractInput,
    validations: [],
    loadContract: loadExampleContract,
  }

  const commandConfigWithSuffixes = {
    ...simpleCommandConfig,
    ...{ ux: { category: 'example', function: 'action', suffixes: ['send', 'bob'], examples: [] } },
  }

  const command = registerExecuteCommand(makeExecuteCommand(simpleCommandConfig))
  const commandWithSuffixes = registerExecuteCommand(makeExecuteCommand(commandConfigWithSuffixes))

  it('Command ID generation', async () => {
    expect(command.id).toEqual('example:action')
    expect(commandWithSuffixes.id).toEqual('example:action:send:bob')
  })

  it('Command input creation', async () => {
    const commandInstance = await command.create({ a: 'a', b: '20' }, [])
    expect(commandInstance.input.user).toEqual({ a: 'a', b: 20 })
    expect(commandInstance.input.contract).toEqual(['a', 20])
  })

  // Deployment succeeds
  it('Command deploy execution', async () => {
    const commandInstance = await command.create({ a: 'a', b: '20' }, [])
  })

  // Exectition with no wallet succeeds
  it('Command no wallet execution', async () => {
    const commandInstance = await command.create({ a: 'a', b: '20' }, [])
  })

  // Execution with account wallet succeeds
  it('Command account execution', async () => {
    const commandInstance = await command.create({ a: 'a', b: '20' }, [])
  })
})
