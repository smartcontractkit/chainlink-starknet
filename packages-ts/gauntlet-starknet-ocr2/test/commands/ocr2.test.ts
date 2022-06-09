import { startNetwork, IntegratedDevnet, makeProvider } from '@chainlink/gauntlet-starknet'
import deployCommand from '../../src/commands/ocr2/deploy'
import setBillingCommand from '../../src/commands/ocr2/setBilling'
import setConfigCommand from '../../src/commands/ocr2/setConfig'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/gauntlet-starknet-example/test/utils'
import { loadContract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

describe('OCR2 Contract', () => {
  let network: IntegratedDevnet
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          minSubmissionValue: 1,
          maxSubmissionValue: 1000,
          decimals: 18,
          name: 'test-feed',
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
    'Set billing with no wallet',
    async () => {
      const command = await registerExecuteCommand(setBillingCommand).create(
        {
          observationPaymentGjuels: 1,
          transmissionPaymentGjuels: 1,
          noWallet: true,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const ocr2 = loadContract(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(ocr2.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ocr2Contract.billing()
      const billing = response[0]
      expect(billing.observation_payment_gjuels.toNumber()).toEqual(1)
      expect(billing.transmission_payment_gjuels.toNumber()).toEqual(1)
    },
    TIMEOUT,
  )

  it(
    'Set default config with no wallet',
    async () => {
      const command = await registerExecuteCommand(setConfigCommand).create(
        {
          default: true,
          noWallet: true,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const ocr2 = loadContract(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(ocr2.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ocr2Contract.transmitters()
      const transmitters = response[0]
      const expected = [
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733',
      ]
      expect(transmitters).toEqual(expected.map((transmitter) => new BN(transmitter.split('x')[1], 16)))
    },
    TIMEOUT,
  )

  it(
    'Set config using --input with no wallet',
    async () => {
      const command = await registerExecuteCommand(setConfigCommand).create(
        {
          noWallet: true,
          input: {
            f: 1,
            signers: [
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737',
            ],
            transmitters: [
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736',
              '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737',
            ],
            onchainConfig: 1,
            offchainConfig: [1],
            offchainConfigVersion: 2,
          },
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const ocr2 = loadContract(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(ocr2.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ocr2Contract.transmitters()
      const transmitters = response[0]
      const expected = [
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737',
      ]
      expect(transmitters).toEqual(expected.map((transmitter) => new BN(transmitter.split('x')[1], 16)))
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
