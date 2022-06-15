import { Account, Contract, defaultProvider, ec } from 'starknet'
import { CallContractResponse } from 'starknet/types'

import { loadContract } from './index'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'OCR2_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/.env' })
async function main() {
  const keyPair = ec.getKeyPair(process.env.PRIVATE_KEY as string)
  account = new Account(defaultProvider, process.env.ACCOUNT_ADDRESS as string, keyPair)
  const OCR2Artifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(OCR2Artifact.abi, process.env.CONSUMER as string)

  const transaction = await account.execute(
    {
      contractAddress: consumer.address,
      entrypoint: 'storeLatestRound',
      calldata: [],
    },
    [consumer.abi],
    { maxFee: 30000000000000 },
  )

  console.log('Waiting for Tx to be Accepted on Starknet...')
  await defaultProvider.waitForTransaction(transaction.transaction_hash)

  const latestRound = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readStoredRound',
    calldata: [],
  })
  printResult(latestRound)
}

function printResult(latestRound: CallContractResponse) {
  console.log('answer= ', parseInt(latestRound.result[0], 16))
  console.log('block_num= ', parseInt(latestRound.result[1], 16))
  console.log('observation_timestamp= ', parseInt(latestRound.result[2], 16))
  console.log('transmission_timestamp= ', parseInt(latestRound.result[3], 16))
}

main()
