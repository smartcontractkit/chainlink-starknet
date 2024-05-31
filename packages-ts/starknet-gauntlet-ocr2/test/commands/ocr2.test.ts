import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/ocr2/deploy'
import setBillingCommand from '../../src/commands/ocr2/setBilling'
import setConfigCommand from '../../src/commands/ocr2/setConfig'
import transferOwnershipCommand from '../../src/commands/ocr2/transferOwnership'
import acceptOwnershipCommand from '../../src/commands/ocr2/acceptOwnership'
import deployACCommand from '../../src/commands/accessController/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  StarknetAccount,
  fetchAccount,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContract } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract, InvokeTransactionReceiptResponse, RpcProvider } from 'starknet'
import { RDDTempFile } from './utils'
import path from 'path'

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
    maxDurationQueryNanoseconds: 2000000000,
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

const getNumCallsPerAddress = (txReceipt: InvokeTransactionReceiptResponse) => {
  const counter = new Map<string, number>()
  txReceipt.events.forEach((ev) => {
    const count = counter.get(ev.from_address)
    if (count == null) {
      counter.set(ev.from_address, 1)
    } else {
      counter.set(ev.from_address, count + 1)
    }
  })
  return counter
}

describe('OCR2 Contract', () => {
  const rddTempFilePath = path.join(__dirname, 'tmp', 'rdd.json')
  const rddTempFile = new RDDTempFile(rddTempFilePath)
  const feedAddresses = new Array<string>()
  const numFeeds = 3

  let owner: StarknetAccount

  beforeAll(async () => {
    // Fetches a prefunded account from the local starknet devnet node
    owner = await fetchAccount()

    // Deploys an access controller
    // TODO: owner can't be 0 anymore
    const deployAccessControllerCmd = await registerExecuteCommand(deployACCommand).create({}, [])
    const deployAccessControllerReport = await deployAccessControllerCmd.execute()
    expect(deployAccessControllerReport.responses[0].tx.status).toEqual('ACCEPTED')
    const accessController = deployAccessControllerReport.responses[0].contract

    // Deploys one feed at a time (if we try to do this in parallel using
    // `Promise.all` / `Promise.allSettled`, then transaction nonce errors
    // will occur)
    for (let i = 0; i < numFeeds; i++) {
      const deployCmd = await registerExecuteCommand(deployCommand).create(
        {
          input: {
            owner: owner.address,
            maxAnswer: 10000,
            minAnswer: 1,
            decimals: 18,
            description: `Test Feed ${i}`,
            billingAccessController: accessController,
            linkToken: '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
          },
        },
        [],
      )
      const deployReport = await deployCmd.execute()
      expect(deployReport.responses[0].tx.status).toEqual('ACCEPTED')
      feedAddresses.push(deployReport.responses[0].contract)
    }

    // Adds dummy billing configs to the temp RDD file
    feedAddresses.forEach((addr, i) => {
      rddTempFile.setBilling(addr, {
        observationPaymentGjuels: String(i),
        transmissionPaymentGjuels: String(i),
        gasBase: String(i),
        gasPerSignature: String(i),
      })
    })

    // Writes the RDD file to disk
    await rddTempFile.writeFile()
  }, TIMEOUT)

  afterAll(async () => {
    await rddTempFile.removeFile()
  })

  it(
    'Set billing',
    async () => {
      // transfer overflow on set billing
      const command = await registerExecuteCommand(setBillingCommand).create(
        {
          input: {
            observationPaymentGjuels: 1,
            transmissionPaymentGjuels: 1,
            gasBase: 14951,
            gasPerSignature: 13,
          },
        },
        feedAddresses,
      )

      // Execute the command
      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      // Validate the response
      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the transaction was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      const txReceipt = receipt as InvokeTransactionReceiptResponse

      // Loads the contract
      const { contract } = loadContract(CONTRACT_LIST.OCR2)

      // Creates a map where each key is an address and each corresponding value is
      // the number of times the address is seen in the transaction receipt events
      const counter = getNumCallsPerAddress(txReceipt)

      // Iterate over the feeds
      for (const feedAddress of feedAddresses) {
        // Checks that the feed was updated
        const ocr2Contract = new Contract(contract.abi, feedAddress, provider)
        const billing = await ocr2Contract.billing()
        expect(billing.observation_payment_gjuels).toEqual(BigInt(1))
        expect(billing.transmission_payment_gjuels).toEqual(BigInt(1))

        // Checks that the correct number of batch calls were made
        expect(counter.get(feedAddress) ?? 0).toEqual(1)
      }
    },
    TIMEOUT,
  )

  it(
    'Set billing using --rdd',
    async () => {
      // transfer overflow on set billing
      const command = await registerExecuteCommand(setBillingCommand).create(
        {
          rdd: rddTempFile.filepath,
        },
        feedAddresses,
      )

      // Execute the command
      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      // Validate the response
      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the transaction was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      const txReceipt = receipt as InvokeTransactionReceiptResponse

      // Loads the contract
      const { contract } = loadContract(CONTRACT_LIST.OCR2)

      // Creates a map where each key is an address and each corresponding value is
      // the number of times the address is seen in the transaction receipt events
      const counter = getNumCallsPerAddress(txReceipt)

      // Iterates over the feeds and checks that the values were updated and that the
      // number of batch calls is correct
      await Promise.all(
        feedAddresses.map(async (feedAddress, i) => {
          const ocr2Contract = new Contract(contract.abi, feedAddress, provider)
          const billing = await ocr2Contract.billing()
          expect(billing.observation_payment_gjuels).toEqual(BigInt(i))
          expect(billing.transmission_payment_gjuels).toEqual(BigInt(i))
          expect(counter.get(feedAddress) ?? 0).toEqual(1)
        }),
      )
    },
    TIMEOUT,
  )

  it(
    'Set config using --input',
    async () => {
      const command = await registerExecuteCommand(setConfigCommand).create(
        {
          input: validInput,
        },
        feedAddresses,
      )

      // Execute the command
      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      // Validate the response
      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the transaction was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      const txReceipt = receipt as InvokeTransactionReceiptResponse

      // Loads the contract
      const { contract } = loadContract(CONTRACT_LIST.OCR2)

      // Creates a map where each key is an address and each corresponding value is
      // the number of times the address is seen in the transaction receipt events
      const counter = getNumCallsPerAddress(txReceipt)

      // Iterate over the feeds
      for (const feedAddress of feedAddresses) {
        // Get a reference to the contract
        const ocr2Contract = new Contract(contract.abi, feedAddress, provider)
        const resultTransmitters = await ocr2Contract.transmitters()

        // retrieve signer keys from transaction event
        // based on event struct: https://github.com/smartcontractkit/chainlink-starknet/blob/develop/contracts/src/chainlink/ocr2/aggregator.cairo#L260
        // TODO: use StarknetContract decodeEvents from starknet-hardhat-plugin instead
        const eventData = txReceipt.events[0].data
        // reconstruct signers array from event
        const eventSigners: bigint[] = []
        for (let i = 0; i < signers.length; i++) {
          const signer = BigInt(eventData[2 + 2 * i]) // split according to event structure
          eventSigners.push(signer)
        }

        // Checks that the feed was updated
        expect(eventSigners).toEqual(
          // eaiser to remove prefix and 0x and then add 0x back
          signers.map((s) => BigInt(`0x${s.replace('ocr2on_starknet_', '').replace('0x', '')}`)),
        ) // remove all prefixes
        expect(resultTransmitters).toEqual(transmitters.map((transmitter) => BigInt(transmitter)))

        // Checks that the correct number of batch calls were made
        expect(counter.get(feedAddress) ?? 0).toEqual(1)
      }
    },
    TIMEOUT,
  )

  it(
    'Transfer ownership',
    async () => {
      const command = await registerExecuteCommand(transferOwnershipCommand).create(
        {
          // Trivially transfer ownership to the same address
          newOwner: owner.address,
        },
        feedAddresses,
      )

      // Execute the command
      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      // Validate the response
      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the transaction was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      const txReceipt = receipt as InvokeTransactionReceiptResponse

      // Creates a map where each key is an address and each corresponding value is
      // the number of times the address is seen in the transaction receipt events
      const counter = getNumCallsPerAddress(txReceipt)

      // Checks that the correct number of batch calls were made
      for (const feedAddress of feedAddresses) {
        expect(counter.get(feedAddress) ?? 0).toEqual(1)
      }
    },
    TIMEOUT,
  )

  it(
    'Accept ownership',
    async () => {
      const command = await registerExecuteCommand(acceptOwnershipCommand).create({}, feedAddresses)

      // Execute the command
      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      // Validate the response
      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the transaction was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      const txReceipt = receipt as InvokeTransactionReceiptResponse

      // Creates a map where each key is an address and each corresponding value is
      // the number of times the address is seen in the transaction receipt events
      const counter = getNumCallsPerAddress(txReceipt)

      // Checks that the correct number of batch calls were made
      for (const feedAddress of feedAddresses) {
        expect(counter.get(feedAddress) ?? 0).toEqual(1)
      }
    },
    TIMEOUT,
  )
})
