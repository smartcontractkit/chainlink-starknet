import { Account, Contract } from 'starknet'
import { createDeployerAccount, loadContract, makeProvider } from './utils'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'Aggregator_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/../.env' })

export async function readDecimals() {
  const provider = makeProvider()
  account = createDeployerAccount(provider)

  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)

  const decimals = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readDecimals',
    calldata: [],
  })

  console.log('decimals= ', parseInt(decimals.result[0], 16))
  return parseInt(decimals.result[0], 16)
}

readDecimals()
