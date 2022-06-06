import { startNetwork, IntegratedDevnet } from '@chainlink/gauntlet-starknet'
import deployCommand from '../../src/commands/account/deploy'
import { registerExecuteCommand, TIMEOUT } from '@chainlink/gauntlet-starknet-example/test/utils'

describe('OZ Account Contract', () => {
  let network: IntegratedDevnet

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})