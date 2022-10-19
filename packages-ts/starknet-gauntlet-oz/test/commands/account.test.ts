import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/account/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import {
  loadContract,
  CONTRACT_LIST,
  calculateAddress,
  equalAddress,
} from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

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

      const oz = loadContract(CONTRACT_LIST.ACCOUNT)
      const ozContract = new Contract(oz.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ozContract.getPublicKey()
      const onChainPubKey = response[0]
      expect(onChainPubKey).toEqual(new BN(publicKey.split('x')[1], 16))
    },
    TIMEOUT,
  )

  it(
    'Deployment with salt',
    async () => {
      let salt: number = 100
      const command = await registerExecuteCommand(deployCommand).create(
        {
          salt,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
      publicKey = report.data.publicKey

      // debugging if this fails:
      // has there been a OZ account contract update? if yes, make sure to update the contract hash for gauntlet and keystore
      let calcAddress = calculateAddress(salt, publicKey)
      expect(equalAddress(calcAddress, contractAddress)).toBe(true)
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
