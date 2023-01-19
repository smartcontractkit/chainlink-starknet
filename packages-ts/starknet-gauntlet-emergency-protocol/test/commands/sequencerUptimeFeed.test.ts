import deployCommand from '../../src/commands/sequencerUptimeFeed/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Sequencer Uptime Feed Contract', () => {
  let network
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  describe('deploy', () => {
    it(
      'Deployment with default input',
      async () => {
        const command = await registerExecuteCommand(deployCommand).create(
          {
            initialStatus: 0,
          },
          [],
        )

        const report = await command.execute()
        expect(report.responses[0].tx.status).toEqual('ACCEPTED')

        contractAddress = report.responses[0].contract
      },
      TIMEOUT,
    )
  })

  afterAll(() => {
    network.stop()
  })
})
