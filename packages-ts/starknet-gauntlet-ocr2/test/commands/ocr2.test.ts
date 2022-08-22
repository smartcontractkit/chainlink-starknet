import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/ocr2/deploy'
import setBillingCommand from '../../src/commands/ocr2/setBilling'
import setConfigCommand from '../../src/commands/ocr2/setConfig'
import deployACCommand from '../../src/commands/accessController/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContractOcr2, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

const signers = [
  'ocr2on_starknet_04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730', // ocr2on_starknet_<key>
  'ocr2on_starknet_0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731', // ocr2on_starknet_0x<key>
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732', // 0x<key>
  '04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733', // <key>
]

const transmitters = [
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737',
]

const validInput = {
  f: 1,
  signers,
  transmitters,
  onchainConfig: '',
  offchainConfig: {
    deltaProgressNanoseconds: 8000000000,
    deltaResendNanoseconds: 30000000000,
    deltaRoundNanoseconds: 3000000000,
    deltaGraceNanoseconds: 500000000,
    deltaStageNanoseconds: 20000000000,
    rMax: 5,
    s: [1, 2],
    offchainPublicKeys: [
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852090',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852091',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852092',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852093',
    ],
    peerIds: signers,
    reportingPluginConfig: {
      alphaReportInfinite: false,
      alphaReportPpb: 0,
      alphaAcceptInfinite: false,
      alphaAcceptPpb: 0,
      deltaCNanoseconds: 0,
    },
    maxDurationQueryNanoseconds: 0,
    maxDurationObservationNanoseconds: 1000000000,
    maxDurationReportNanoseconds: 200000000,
    maxDurationShouldAcceptFinalizedReportNanoseconds: 200000000,
    maxDurationShouldTransmitAcceptedReportNanoseconds: 200000000,
    configPublicKeys: [
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852094',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852095',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852096',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852097',
    ],
  },
  offchainConfigVersion: 2,
  secret: 'awe accuse polygon tonic depart acuity onyx inform bound gilbert expire',
}

describe('OCR2 Contract', () => {
  let network: IntegratedDevnet
  let contractAddress: string
  let accessController: string

  beforeAll(async () => {
    network = await startNetwork()

    const command = await registerExecuteCommand(deployACCommand).create({}, [])

    const report = await command.execute()
    accessController = report.responses[0].contract
  }, TIMEOUT)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          input: {
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

      const ocr2 = loadContractOcr2(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(ocr2.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const response = await ocr2Contract.billing()
      const billing = response[0]
      expect(billing.observation_payment_gjuels.toNumber()).toEqual(1)
      expect(billing.transmission_payment_gjuels.toNumber()).toEqual(1)
    },
    TIMEOUT,
  )

  it(
    'Set config using --input with no wallet',
    async () => {
      const command = await registerExecuteCommand(setConfigCommand).create(
        {
          noWallet: true,
          input: validInput,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const provider = makeProvider(LOCAL_URL).provider
      const ocr2 = loadContractOcr2(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(ocr2.abi, contractAddress, provider)
      const response = await ocr2Contract.transmitters()
      const resultTrasmitters = response[0]

      // retrieve signer keys from transaction event
      // based on event struct: https://github.com/smartcontractkit/chainlink-starknet/blob/develop/contracts/contracts/ocr2/aggregator.cairo#L260
      const trace = await provider.getTransactionTrace(report.responses[0].tx.hash)
      const eventData = trace.function_invocation.events[0].data
      // reconstruct signers array from event
      let eventSigners = []
      for (let i = 0; i < signers.length; i++) {
        const signer = new BN(eventData[4 + 2 * i].replace('0x', ''), 16) // split according to event structure
        eventSigners.push(signer)
      }

      expect(eventSigners).toEqual(signers.map((s) => new BN(s.replace('ocr2on_starknet_', '').replace('0x', ''), 16))) // remove all prefixes
      expect(resultTrasmitters).toEqual(transmitters.map((transmitter) => new BN(transmitter.split('x')[1], 16)))
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
