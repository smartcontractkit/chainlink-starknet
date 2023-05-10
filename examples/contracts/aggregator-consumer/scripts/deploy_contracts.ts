import { Contract } from 'starknet'
import { loadContract, createDeployerAccount, loadContractPath, makeProvider } from './utils'
import fs from 'fs'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'Aggregator_consumer'

const MOCK_NAME = 'MockAggregator'

const PRICE_CONSUMER_NAME = 'Price_Consumer_With_Sequencer_Check'

const UPTIME_FEED_PATH =
  '../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed'
const UPTIME_FEED_NAME = 'sequencer_uptime_feed'

const DECIMALS = '18'

dotenv.config({ path: __dirname + '/../.env' })

export async function deployContract() {
  const provider = makeProvider()
  const MockArtifact = loadContract(MOCK_NAME)
  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  const priceConsumerArtifact = loadContract(PRICE_CONSUMER_NAME)
  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH, UPTIME_FEED_NAME)

  const account = createDeployerAccount(provider)

  const declareDeployMock = await account.declareAndDeploy({
    contract: MockArtifact,
    constructorCalldata: [DECIMALS],
  })

  const mockDeploy = new Contract(
    MockArtifact.abi,
    declareDeployMock.deploy.contract_address,
    provider,
  )

  const declareDeployAggregator = await account.declareAndDeploy({
    contract: AggregatorArtifact,
    constructorCalldata: [mockDeploy.address as string],
  })

  const consumerDeploy = new Contract(
    AggregatorArtifact.abi,
    declareDeployAggregator.deploy.contract_address,
    provider,
  )

  const declareDeployUptimeFeed = await account.declareAndDeploy({
    contract: UptimeFeedArtifact,
    constructorCalldata: ['0', account.address],
  })

  const uptimeFeedDeploy = new Contract(
    UptimeFeedArtifact.abi,
    declareDeployUptimeFeed.deploy.contract_address,
    provider,
  )

  const declareDeployPriceConsumer = await account.declareAndDeploy({
    contract: priceConsumerArtifact,
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
  fs.appendFile(__dirname + '/../.env', '\nUPTIME_FEED=' + uptimeFeedDeploy.address, function (
    err,
  ) {
    if (err) throw err
  })
}

deployContract()
