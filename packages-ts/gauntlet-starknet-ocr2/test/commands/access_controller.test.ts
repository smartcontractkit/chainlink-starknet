import { startNetwork, IntegratedDevnet } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/accessController/deploy'
import { registerExecuteCommand, TIMEOUT } from '@chainlink/starknet-gauntlet-example/test/utils'

describe('Access Controller Contract', () => {
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
