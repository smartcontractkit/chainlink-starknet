import { starknet } from 'hardhat'
import dotenv from 'dotenv'

const PRICE_CONSUMER_NAME = 'Price_Consumer'

dotenv.config({ path: __dirname + '/.env' })

async function main() {
  const account = await starknet.getAccountFromAddress(
    process.env.ACCOUNT_ADDRESS as string,
    process.env.PRIVATE_KEY as string,
    'OpenZeppelin',
  )

  const priceConsumerFactory = await starknet.getContractFactory(PRICE_CONSUMER_NAME)
  const priceConsumerDeploy = await priceConsumerFactory.deploy({
    uptime_feed_address: process.env.UPTIME_FEED,
    aggregator_address: process.env.MOCK_AGGREGATOR,
  })

  const latestPrice = await account.call(priceConsumerDeploy, 'get_latest_price')
  console.log('answer= ', latestPrice)
}

main()
