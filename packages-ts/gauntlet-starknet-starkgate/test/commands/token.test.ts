// TODO: Import from package
import deployOZCommand from '../../../gauntlet-starknet-oz/src/commands/account/deploy'
import deployCommand from '../../src/commands/token/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/gauntlet-starknet/test/utils'

describe('Token Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string
  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      account = report.responses[0].contract
      privateKey = report.data.privateKey
    },
    TIMEOUT,
  )

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          link: true,
          account: account,
          pk: privateKey,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      tokenContractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  // TODO: Test mint

  // TODO: Test transfer

  afterAll(() => {
    network.stop()
  })
})
