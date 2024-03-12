import deployOCR2Command from '../../src/commands/ocr2/deploy'
import deployACCommand from '../../src/commands/accessController/deploy'
import deployProxyCommand from '../../src/commands/proxy/deploy'
import proposeAggregatorCommand from '../../src/commands/proxy/proposeAggregator'
import confirmAggregatorCommand from '../../src/commands/proxy/confirmAggregator'
import {
  StarknetAccount,
  fetchAccount,
  registerExecuteCommand,
  TIMEOUT,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Proxy Contract', () => {
  let account: StarknetAccount
  let contractAddress: string
  let accessController: string
  let proxy: string

  beforeAll(async () => {
    account = await fetchAccount()
  })

  it(
    'Deploy AC',
    async () => {
      // TODO: owner can't be 0 anymore
      const command = await registerExecuteCommand(deployACCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      accessController = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Deploy OCR2',
    async () => {
      const command = await registerExecuteCommand(deployOCR2Command).create(
        {
          input: {
            owner: account.address,
            maxAnswer: 10000,
            minAnswer: 1,
            decimals: 18,
            description: 'Test Feed',
            billingAccessController: accessController,
            linkToken: '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
          },
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Deploy Proxy',
    async () => {
      const command = await registerExecuteCommand(deployProxyCommand).create(
        {
          input: {
            owner: account.address,
            address: contractAddress,
          },
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      proxy = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Propose Proxy Aggregator',
    async () => {
      const command = await registerExecuteCommand(proposeAggregatorCommand).create(
        {
          input: {
            owner: account.address,
            address: contractAddress,
          },
        },
        [proxy],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  it(
    'Confirm Proxy Aggregator',
    async () => {
      const command = await registerExecuteCommand(confirmAggregatorCommand).create(
        {
          input: {
            owner: account.address,
            address: contractAddress,
          },
        },
        [proxy],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )
})
