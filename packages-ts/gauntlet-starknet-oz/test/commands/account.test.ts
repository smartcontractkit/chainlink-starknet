import { startNetwork, IntegratedDevnet, makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/account/deploy'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/starknet-gauntlet-example/test/utils'
import { loadContract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

describe('OZ Account Contract', () => {
  let network: IntegratedDevnet
  let publicKey: string
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
      publicKey = report.data.publicKey

      const oz = loadContract(CONTRACT_LIST.ACCOUNT)
      const ozContract = new Contract(oz.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ozContract.get_public_key()
      const onChainPubKey = response[0]
      expect(onChainPubKey).toEqual(new BN(publicKey.split('x')[1], 16))
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
