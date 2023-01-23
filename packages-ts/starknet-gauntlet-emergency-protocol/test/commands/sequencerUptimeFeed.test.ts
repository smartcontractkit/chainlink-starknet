import deployCommand from '../../src/commands/sequencerUptimeFeed/deploy'
import setL1Sender from '../../src/commands/sequencerUptimeFeed/setL1Sender'
import getLatestRoundData from '../../src/commands/sequencerUptimeFeed/inspection/inspection'
import {
  registerExecuteCommand,
  registerInspectCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Sequencer Uptime Feed Contract', () => {
  let network: IntegratedDevnet
  let uptimeFeedAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  // deploy uptime feed contract
  beforeEach(async () => {
    const command = await registerExecuteCommand(deployCommand).create(
      {
        initialStatus: 0,
      },
      [],
    )

    const report = await command.execute()
    expect(report.responses[0].tx.status).toEqual('ACCEPTED')

    uptimeFeedAddress = report.responses[0].contract
  }, TIMEOUT)

  describe('setL1Sender', () => {
    it(
      'Sets with a non-Zero valid eth address',
      async () => {
        const command = await registerExecuteCommand(setL1Sender).create(
          {
            address: '0x0000000000000000000000000000000000000007',
          },
          [uptimeFeedAddress],
        )

        const report = await command.execute()
        expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      },
      TIMEOUT,
    )
  })

  describe('inspect uptime feed status', () => {
    it('Retrieves 0', async () => {
      const inspectCommand = await registerInspectCommand(getLatestRoundData).create({}, [
        uptimeFeedAddress,
      ])

      const report = await inspectCommand.execute()
      expect(report.data.data.latest_round_data.answer).toEqual('0')
    })
  })

  afterAll(() => {
    network.stop()
  })
})
