import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { Contract } from 'starknet'
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
import { IntegratedDevnet, startNetwork } from '../utils/network'

const TIMEOUT = 100000
const LOCAL_URL = 'http://127.0.0.1:5050/'

const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: (flags) => {
      return {
        providerUrl: LOCAL_URL,
        pk: flags.pk,
        account: flags.account,
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
    contractId: '',
    category: 'example',
    action: 'action',
    ux: {
      description: '',
      examples: [],
    },
    makeUserInput,
    makeContractInput,
    validations: [],
    loadContract: loadExampleContract,
  }

  const commandConfigWithSuffixes = {
    ...simpleCommandConfig,
    ...{ suffixes: ['send', 'bob'] },
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
})

describe('Execute with network', () => {
  let network: IntegratedDevnet
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  it(
    'Command deploy execution',
    async () => {
      const makeUserInput = async (flags, args) => {
        return
      }

      const makeContractInput = async (userInput) => {
        return {}
      }

      const deployCommandConfig: ExecuteCommandConfig<any, any> = {
        contractId: '',
        category: 'example',
        action: 'deploy',
        ux: {
          description: '',
          examples: [],
        },
        makeUserInput,
        makeContractInput,
        validations: [],
        loadContract: loadExampleContract,
      }

      const command = registerExecuteCommand(makeExecuteCommand(deployCommandConfig))

      const commandInstance = await command.create({}, [])
      const report = await commandInstance.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Command no wallet execution',
    async () => {
      const makeUserInput = async (flags, args) => {
        return
      }

      const makeContractInput = async (userInput) => {
        return [100]
      }

      const increaseCommandConfig: ExecuteCommandConfig<any, any> = {
        contractId: '',
        category: 'example',
        action: 'increase_balance',
        ux: {
          description: '',
          examples: [],
        },
        makeUserInput,
        makeContractInput,
        validations: [],
        loadContract: loadExampleContract,
      }

      const command = registerExecuteCommand(makeExecuteCommand(increaseCommandConfig))

      const commandInstance = await command.create({ noWallet: true }, [contractAddress])
      const report = await commandInstance.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const contract = new Contract(loadExampleContract().abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const balance = await contract.get_balance()

      expect(new BN(balance.res).toString()).toEqual('100')
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
