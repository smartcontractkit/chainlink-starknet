import { Contract, Account, CallContractResponse, Result } from 'starknet'

import { createDeployerAccount, loadContract, makeProvider } from './utils'
import dotenv from 'dotenv'

const CONTRACT_NAME = 'AggregatorConsumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/../.env' })

async function readContinuously() {
  const provider = makeProvider()

  account = createDeployerAccount(provider)

  const AggregatorArtifact = loadContract(CONTRACT_NAME)

  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)
  consumer.connect(account)
  setInterval(callFunction, 3000)
}

async function callFunction() {
  const latestRound = await consumer.call('read_latest_round')
  const decimals = await consumer.call('read_decimals')
  printResult(latestRound, decimals)
}

function printResult(latestRound: Result, decimals: Result) {
  console.log('---------------')
  console.log('round_id= ', latestRound['round_id'])
  console.log('answer= ', latestRound['answer'])
  console.log('block_num= ', latestRound['block_num'])
  console.log('started_at= ', latestRound['started_at'])
  console.log('updated_at= ', latestRound['updated_at'])
  console.log('decimals= ', decimals.toString())
}

readContinuously()
