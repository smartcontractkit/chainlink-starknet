import dotenv from 'dotenv'
import { createDeployerAccount, loadContract, loadContractPath, makeProvider } from './utils'
import { CompiledContract, Contract, GatewayError } from 'starknet'

const PRICE_CONSUMER_NAME = 'AggregatorPriceConsumerWithSequencer'
const UPTIME_FEED_PATH = '../../../../contracts/target/release/chainlink_SequencerUptimeFeed.sierra'

dotenv.config({ path: __dirname + '/../.env' })

const SEQ_UP_REPORT_STALE = 'L2 seq up & report stale'
const SEQ_DOWN_REPORT_STALE = 'L2 seq down & report stale'
const SEQ_DOWN_REPORT_OK = 'L2 seq down & report ok'

const HEX_SEQ_UP_REPORT_STALE = revertMessageHex(SEQ_UP_REPORT_STALE)
const HEX_SEQ_DOWN_REPORT_STALE = revertMessageHex(SEQ_DOWN_REPORT_STALE)
const HEX_SEQ_DOWN_REPORT_OK = revertMessageHex(SEQ_DOWN_REPORT_OK)

function revertMessageHex(msg: string): string {
  return Buffer.from(msg, 'utf8').toString('hex')
}

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
  } catch (e) {
    // transaction reverted because sequencer is down or report is stale
    console.log('Getting latest price not possible (reason below)')
    if (e instanceof GatewayError) {
      switch (true) {
        case e.message.includes(HEX_SEQ_UP_REPORT_STALE): {
          console.log(SEQ_UP_REPORT_STALE)
          break
        }
        case e.message.includes(HEX_SEQ_DOWN_REPORT_STALE): {
          console.log(SEQ_DOWN_REPORT_STALE)
          break
        }
        case e.message.includes(HEX_SEQ_DOWN_REPORT_OK): {
          console.log(SEQ_DOWN_REPORT_OK)
          break
        }
        default:
          console.log(e)
          break
      }
    } else {
      console.log(e)
    }
  }
}

getLatestPrice()
