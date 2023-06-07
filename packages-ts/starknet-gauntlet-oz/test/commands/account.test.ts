import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/account/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { accountContractLoader, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'

describe('OZ Account Contract', () => {
  let network: IntegratedDevnet
  let publicKey: string
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, TIMEOUT)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
      publicKey = report.data.publicKey

      const { contract: oz } = accountContractLoader()
      const ozContract = new Contract(oz.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const { publicKey: onChainPubKey } = await ozContract.getPublicKey()
      expect(onChainPubKey).toEqual(BigInt(publicKey))
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
