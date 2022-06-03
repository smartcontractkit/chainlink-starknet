import { startNetwork, IntegratedDevnet } from '@chainlink/gauntlet-starknet'
import deployCommand from '../../src/commands/accessController/deploy'
import { registerExecuteCommand, TIMEOUT } from '@chainlink/gauntlet-starknet-example/test/utils'

describe('Access Controller Contract', () => {
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
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
