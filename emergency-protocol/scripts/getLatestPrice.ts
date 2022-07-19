import { Account, Contract, defaultProvider, ec } from 'starknet'
import { CallContractResponse } from 'starknet/types'
import { starknet } from 'hardhat'
import { loadContract } from './index'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'Price_Consumer'
// let account: Account

dotenv.config({ path: __dirname + '/.env' })

async function main() {

  const account = await starknet.getAccountFromAddress(process.env.ACCOUNT_ADDRESS as string, process.env.PRIVATE_KEY as string, "OpenZeppelin");

  const ConsumerFeedFactory = await starknet.getContractFactory(CONSUMER_NAME)
  const ConsumerFeedDeploy = await ConsumerFeedFactory.deploy({ uptime_feed_address: process.env.UPTIME_FEED, aggregator_address: process.env.MOCK_AGGREGATOR})

  await account.invoke(
    ConsumerFeedDeploy,
    'check_sequencer_state',
  )

  const latestPrice = await account.call(
    ConsumerFeedDeploy,
    'get_latest_price',
  )

  console.log('answer= ', latestPrice)
}

main()