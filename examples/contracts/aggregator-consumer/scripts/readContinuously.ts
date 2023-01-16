import { Contract, Account, CallContractResponse } from 'starknet'

import { createDeployerAccount, loadContract, makeProvider } from './utils'
import dotenv from 'dotenv'

const CONTRACT_NAME = 'Aggregator_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/../.env' })

async function readContinuously() {
  const provider = makeProvider()

  account = createDeployerAccount(provider)

  const AggregatorArtifact = loadContract(CONTRACT_NAME)

  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)
  setInterval(callFunction, 30000)
}

async function callFunction() {
  const latestRound = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readLatestRound',
    calldata: [],
  })

  const decimals = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readDecimals',
    calldata: [],
  })
  printResult(latestRound, decimals)
}

function printResult(latestRound: CallContractResponse, decimals: CallContractResponse) {
  console.log('round_id= ', parseInt(latestRound.result[0], 16))
  console.log('answer= ', parseInt(latestRound.result[1], 16))
  console.log('block_num= ', parseInt(latestRound.result[2], 16))
  console.log('staerted_at= ', parseInt(latestRound.result[3], 16))
  console.log('updated_at= ', parseInt(latestRound.result[4], 16))
  console.log('decimals= ', parseInt(decimals.result[0], 16))
}

readContinuously()
