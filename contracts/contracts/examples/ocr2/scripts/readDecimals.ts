import { Account, Contract, defaultProvider, ec } from 'starknet'
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

  const decimals = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readDecimals',
    calldata: [],
  })

  console.log('decimals= ', parseInt(decimals.result[0], 16))
}

main()
