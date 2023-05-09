import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/account/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContract, CONTRACT_LIST, equalAddress } from '../../src/lib/contracts'
import { Contract } from 'starknet'

describe('OZ Account Contract', () => {
  let network: IntegratedDevnet
  let publicKey: string
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
      publicKey = report.data.publicKey

      const { contract: oz } = loadContract(CONTRACT_LIST.ACCOUNT)
      const ozContract = new Contract(oz.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const onChainPubKey = await ozContract.get_public_key()
      expect(onChainPubKey).toEqual(BigInt(publicKey))
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
