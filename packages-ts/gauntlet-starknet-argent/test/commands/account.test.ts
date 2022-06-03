import { startNetwork, IntegratedDevnet, makeProvider } from '@chainlink/gauntlet-starknet'
import deployCommand from '../../src/commands/account/deploy'
import initCommand from '../../src/commands/account/initialize'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/gauntlet-starknet-example/test/utils'
import { loadContract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

describe('Argent Account Contract', () => {
  let network: IntegratedDevnet
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Initialization',
    async () => {
      const command = await registerExecuteCommand(initCommand).create(
        {
          noWallet: true,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const publicKey = new BN(report.data.publicKey.split('x')[1], 16)
      const account = loadContract(CONTRACT_LIST.ACCOUNT)
      const accountContract = new Contract(account.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await accountContract['get_signer']()
      const signer = response[0]
      expect(signer).toEqual(publicKey)
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
