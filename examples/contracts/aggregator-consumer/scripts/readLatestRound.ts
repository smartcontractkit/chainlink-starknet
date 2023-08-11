import { Account, Contract, Result } from 'starknet'

import { createDeployerAccount, loadContract, makeProvider } from './utils'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'AggregatorConsumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/../.env' })

export async function readLatestRound() {
  const provider = makeProvider()
  account = createDeployerAccount(provider)

  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)

  consumer.connect(account)

  const latestRound = await consumer.call('read_latest_round')

  printResult(latestRound)
  return latestRound
}

function printResult(latestRound: Result) {
  console.log('round_id= ', latestRound['round_id'])
  console.log('answer= ', latestRound['answer'])
  console.log('block_num= ', latestRound['block_num'])
  console.log('started_at= ', latestRound['started_at'])
  console.log('updated_at= ', latestRound['updated_at'])
}

readLatestRound()
