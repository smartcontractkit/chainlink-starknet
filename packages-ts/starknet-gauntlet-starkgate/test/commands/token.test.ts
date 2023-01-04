import deployCommand from '../../src/commands/token/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Token Contract', () => {
  let network: IntegratedDevnet
  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          link: true,
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
