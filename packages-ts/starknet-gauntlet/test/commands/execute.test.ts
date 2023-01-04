import { ExecuteCommandConfig, makeExecuteCommand } from '../../src/index'
import {
  devnetAccount0Address,
  devnetPrivateKey,
  loadExampleContract,
  registerExecuteCommand,
} from '../utils'
import { IntegratedDevnet, startNetwork } from '../utils/network'

const TIMEOUT = 100000

let account: string = devnetAccount0Address
let privateKey: string = devnetPrivateKey

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
    network = await startNetwork({ seed: 0 })
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

  afterAll(() => {
    network.stop()
  })
})
