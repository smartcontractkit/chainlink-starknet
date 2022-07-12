import { Account, Contract, defaultProvider, ec } from 'starknet'
import { CallContractResponse } from 'starknet/types'

import { loadContract } from './index'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'Aggregator_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/.env' })

async function main() {
  const keyPair = ec.getKeyPair(process.env.PRIVATE_KEY as string)
  account = new Account(defaultProvider, process.env.ACCOUNT_ADDRESS as string, keyPair)
  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)

  const latestRound = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readLatestRound',
    calldata: [],
  })
  printResult(latestRound)
}

function printResult(latestRound: CallContractResponse) {
  console.log('\nround_id= ', parseInt(latestRound.result[0], 16))
  console.log('answer= ', parseInt(latestRound.result[1], 16))
  console.log('block_num= ', parseInt(latestRound.result[2], 16))
  console.log('observation_timestamp= ', parseInt(latestRound.result[3], 16))
  console.log('transmission_timestamp= ', parseInt(latestRound.result[4], 16))
}

main()
