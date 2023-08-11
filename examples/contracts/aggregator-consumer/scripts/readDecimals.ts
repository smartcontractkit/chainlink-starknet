import { Account, Contract } from 'starknet'
import { createDeployerAccount, loadContract, makeProvider } from './utils'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'AggregatorConsumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/../.env' })

export async function readDecimals() {
  const provider = makeProvider()
  account = createDeployerAccount(provider)

  const AggregatorArtifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(AggregatorArtifact.abi, process.env.CONSUMER as string)

  consumer.connect(account)

  const decimals = await consumer.call('read_decimals')

  console.log('decimals= ', decimals.toString())
  return decimals
}

readDecimals()
