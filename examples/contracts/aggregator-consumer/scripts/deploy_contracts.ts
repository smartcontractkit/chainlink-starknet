import { Contract } from 'starknet'
import { loadContract, createDeployerAccount, loadContractPath, makeProvider } from './utils'
import fs from 'fs'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'Aggregator_consumer'
const CONSUMER_CLASS_HASH = '0x68600b3eaf7ed249a0a70fae5c1e2d33f61ab403e4ac89b371f073a6c1c3c64'

const MOCK_NAME = 'MockAggregator'
const MOCK_CLASS_HASH = '0x21d95032de10675f3814955ef5cae12d3daa717955b54f600888111e56779d3'

const PRICE_CONSUMER_NAME = 'Price_Consumer_With_Sequencer_Check'
const PRICE_CONSUMER_CLASS_HASH =
  '0x401d4cc6d7018330e200bca23dfa7ff5199af0f19b81f7d3ca9c2c0ce3fd892'

const UPTIME_FEED_PATH =
  '../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed'
const UPTIME_FEED_NAME = 'sequencer_uptime_feed'
const UPTIME_FEED_HASH = '0x149c2a0a58155a8a091ded689a5fdcccd596bb0c74b301f678c1b46b690f432'

const DECIMALS = '18'

dotenv.config({ path: __dirname + '/../.env' })

export async function deployContract() {
  const provider = makeProvider()
  const MockArtifact = loadContract(MOCK_NAME)
  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  const priceConsumerArtifact = loadContract(PRICE_CONSUMER_NAME)
  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH, UPTIME_FEED_NAME)

  const account = createDeployerAccount(provider)

  const declareDeployMock = await account.declareDeploy({
    contract: MockArtifact,
    classHash: MOCK_CLASS_HASH,
    constructorCalldata: [DECIMALS],
  })

  const mockDeploy = new Contract(
    MockArtifact.abi,
    declareDeployMock.deploy.contract_address,
    provider,
  )

  const declareDeployAggregator = await account.declareDeploy({
    contract: AggregatorArtifact,
    classHash: CONSUMER_CLASS_HASH,
    constructorCalldata: [mockDeploy.address as string],
  })

  const consumerDeploy = new Contract(
    AggregatorArtifact.abi,
    declareDeployAggregator.deploy.contract_address,
    provider,
  )

  const declareDeployUptimeFeed = await account.declareDeploy({
    contract: UptimeFeedArtifact,
    classHash: UPTIME_FEED_HASH,
    constructorCalldata: ['0', account.address],
  })

  const uptimeFeedDeploy = new Contract(
    UptimeFeedArtifact.abi,
    declareDeployUptimeFeed.deploy.contract_address,
    provider,
  )

  const declareDeployPriceConsumer = await account.declareDeploy({
    contract: priceConsumerArtifact,
    classHash: PRICE_CONSUMER_CLASS_HASH,
    constructorCalldata: [uptimeFeedDeploy.address as string, mockDeploy.address as string],
  })

  const priceConsumerDeploy = new Contract(
    priceConsumerArtifact.abi,
    declareDeployPriceConsumer.deploy.contract_address,
    provider,
  )

  fs.appendFile(__dirname + '/../.env', '\nCONSUMER=' + consumerDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/../.env', '\nMOCK=' + mockDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(
    __dirname + '/../.env',
    '\nPRICE_CONSUMER=' + priceConsumerDeploy.address,
    function (err) {
      if (err) throw err
    },
  )
  fs.appendFile(
    __dirname + '/../.env',
    '\nUPTIME_FEED=' + uptimeFeedDeploy.address,
    function (err) {
      if (err) throw err
    },
  )
}

deployContract()
