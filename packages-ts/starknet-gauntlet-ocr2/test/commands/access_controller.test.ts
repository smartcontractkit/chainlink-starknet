// TODO: Import from package
import deployOZCommand from '../../../starknet-gauntlet-oz/src/commands/account/deploy'
import deployCommand from '../../src/commands/accessController/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Access Controller Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

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
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          account: account,
          pk: privateKey,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
