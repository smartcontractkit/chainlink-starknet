import deployCommand from '../../src/commands/accessController/deploy'
import { registerExecuteCommand, TIMEOUT } from '@chainlink/starknet-gauntlet/test/utils'

describe('Access Controller Contract', () => {
  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )
})
