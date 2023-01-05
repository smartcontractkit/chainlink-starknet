import dotenv from 'dotenv'
import { createDeployerAccount, loadContract, loadContractPath, makeProvider } from './utils'
import { Contract } from 'starknet'

const PRICE_CONSUMER_NAME = 'Price_Consumer_With_Sequencer_Check'
const UPTIME_FEED_PATH =
  '../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed'
const UPTIME_FEED_NAME = 'sequencer_uptime_feed'

dotenv.config({ path: __dirname + '/../.env' })

export async function getLatestPrice() {
  const provider = makeProvider()

  const account = createDeployerAccount(provider)

  const priceConsumerArtifact = loadContract(PRICE_CONSUMER_NAME)
  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH, UPTIME_FEED_NAME)

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

  const transaction = await account.execute(
    {
      contractAddress: uptimeFeed.address,
      entrypoint: 'add_access',
      calldata: [priceConsumer.address],
    },
    [UptimeFeedArtifact.abi],
  )

  console.log('Waiting for Tx to be Accepted on Starknet...')
  await provider.waitForTransaction(transaction.transaction_hash)

  const lat = await uptimeFeed.call('latest_round_data')
  const latestPrice = await account.callContract({
    contractAddress: priceConsumer.address,
    entrypoint: 'get_latest_price',
    calldata: [],
  })

  console.log('answer= ', latestPrice)
  return latestPrice
}

getLatestPrice()
