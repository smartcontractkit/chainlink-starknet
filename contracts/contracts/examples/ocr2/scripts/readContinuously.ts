import { Contract, Account, defaultProvider, ec } from 'starknet'
import { CallContractResponse } from 'starknet/types'
import { loadContract } from './index'
import dotenv from 'dotenv'

const CONTRACT_NAME = 'OCR2_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/.env' })

async function main() {
  const keyPair = ec.getKeyPair(process.env.PRIVATE_KEY as string)
  account = new Account(defaultProvider, process.env.ACCOUNT_ADDRESS as string, keyPair)

  const OCR2Artifact = loadContract(CONTRACT_NAME)

  consumer = new Contract(OCR2Artifact.abi, process.env.CONSUMER as string)
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
  console.log('observation_timestamp= ', parseInt(latestRound.result[3], 16))
  console.log('transmission_timestamp= ', parseInt(latestRound.result[4], 16))
  console.log('decimals= ', parseInt(decimals.result[0], 16))
}

main()
