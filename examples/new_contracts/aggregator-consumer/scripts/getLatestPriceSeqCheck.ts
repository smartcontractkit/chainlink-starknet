import dotenv from 'dotenv'
import { createDeployerAccount, loadContract, loadContractPath, makeProvider } from './utils'
import { CompiledContract, Contract, GatewayError } from 'starknet'

const PRICE_CONSUMER_NAME = 'AggregatorPriceConsumerWithSequencer'
const UPTIME_FEED_PATH = '../../../../contracts/target/release/chainlink_SequencerUptimeFeed.sierra'

dotenv.config({ path: __dirname + '/../.env' })

const SEQUENCER_STALE = Buffer.from('L2 seq up & report stale', 'utf8').toString('hex')
console.log(SEQUENCER_STALE, 'sequencer stale msg')

export async function getLatestPrice() {
  const provider = makeProvider()

  const account = createDeployerAccount(provider)

  const priceConsumerArtifact = loadContract(PRICE_CONSUMER_NAME)
  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH) as CompiledContract

  const priceConsumer = new Contract(
    priceConsumerArtifact.abi,
    process.env.PRICE_CONSUMER as string,
    provider,
  )
  const uptimeFeed = new Contract(
    UptimeFeedArtifact.abi,
    process.env.UPTIME_FEED as string,
    provider,
  )

  priceConsumer.connect(account)
  uptimeFeed.connect(account)

  let res = await uptimeFeed.invoke('add_access', [priceConsumer.address])
  console.log('Waiting for add_access Tx to be Accepted on Starknet...')
  await provider.waitForTransaction(res.transaction_hash)

  try {
    console.log('Waiting to get latest price')
    const latestPrice = await priceConsumer.call('get_latest_price', [])
    console.log('answer= ', latestPrice)
    return latestPrice
  } catch (e) {
    // this will occur if at least 60 seconds have elapsed
    // see AggregatorPriceConsumerWithSequencer for more info
    if (e instanceof GatewayError && e.message.includes(SEQUENCER_STALE)) {
      console.log('sequencer is up but stale so the call is reverted')
    }

  }

}

getLatestPrice()
