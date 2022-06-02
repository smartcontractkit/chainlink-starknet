import { startNetwork, IntegratedDevnet } from '@chainlink/gauntlet-starknet'
import deploy from '../../../src/commands/ocr2/deploy'
import { registerExecuteCommand, TIMEOUT } from '../../../../gauntlet-starknet-example/test/utils'

describe('OCR2 Contract Deployment', () => {
  let network: IntegratedDevnet
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deployment with default input',
    async () => {
      const command = await registerExecuteCommand(deploy).create({
        minSubmissionValue: 1,
        maxSubmissionValue: 1000,
        decimals: 18,
        name: "test-feed"
      }, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      
      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
