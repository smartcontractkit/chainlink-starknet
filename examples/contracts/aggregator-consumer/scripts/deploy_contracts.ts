import { CairoAssembly, CallData, CompiledContract, Contract } from 'starknet'
import { loadContract, createDeployerAccount, loadContractPath, makeProvider, loadCasmContract } from './utils'
import * as fs from 'fs'
import * as dotenv from 'dotenv'

const AGGREGATOR = 'MockAggregator'

const AGGREGATOR_PATH = '../../../../contracts/target/release/chainlink_MockAggregator'

const CONSUMER = 'AggregatorConsumer'

const UPTIME_FEED_PATH = '../../../../contracts/target/release/chainlink_SequencerUptimeFeed'

const PRICE_CONSUMER = 'AggregatorPriceConsumerWithSequencer'

const DECIMALS = '18'

dotenv.config({ path: __dirname + '/../.env' })

export async function deployContract() {
  const provider = makeProvider()
  const AggregatorArtifact = loadContractPath(`${AGGREGATOR_PATH}.sierra`) as CompiledContract
  const ConsumerArtifact = loadContract(CONSUMER)
  const PriceConsumerArtifact = loadContract(PRICE_CONSUMER)
  const UptimeFeedArtifact = loadContractPath(`${UPTIME_FEED_PATH}.sierra`) as CompiledContract

  const account = createDeployerAccount(provider)

  console.log("Deploying Contracts...(this may take 3-5 minutes)")

  const declareDeployAggregator = await account.declareAndDeploy({
    casm: loadContractPath(`${AGGREGATOR_PATH}.casm`) as CairoAssembly,
    contract: AggregatorArtifact,
    constructorCalldata: [DECIMALS],
  })

  const aggregatorDeploy = new Contract(
    AggregatorArtifact.abi,
    declareDeployAggregator.deploy.contract_address,
    provider,
  )

  const declareDeployConsumer = await account.declareAndDeploy({
    casm: loadCasmContract(CONSUMER),
    contract: ConsumerArtifact,
    constructorCalldata: [aggregatorDeploy.address as string],
  })

  const consumerDeploy = new Contract(
    ConsumerArtifact.abi,
    declareDeployConsumer.deploy.contract_address,
    provider,
  )

  const declareDeployUptimeFeed = await account.declareAndDeploy({
    casm: loadContractPath(`${UPTIME_FEED_PATH}.casm`) as CairoAssembly,
    contract: UptimeFeedArtifact,
    constructorCalldata: ['0', account.address],
  })

  const uptimeFeedDeploy = new Contract(
    UptimeFeedArtifact.abi,
    declareDeployUptimeFeed.deploy.contract_address,
    provider,
  )

  const declareDeployPriceConsumer = await account.declareAndDeploy({
    casm: loadCasmContract(PRICE_CONSUMER),
    contract: PriceConsumerArtifact,
    constructorCalldata: [uptimeFeedDeploy.address as string, aggregatorDeploy.address as string],
  })

  const priceConsumerDeploy = new Contract(
    PriceConsumerArtifact.abi,
    declareDeployPriceConsumer.deploy.contract_address,
    provider,
  )

  fs.appendFile(__dirname + '/../.env', '\nCONSUMER=' + consumerDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/../.env', '\nMOCK=' + aggregatorDeploy.address, function (err) {
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
