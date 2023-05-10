import { Account, CompiledContract, Contract, Provider, ec, number } from 'starknet'
import { loadContract, loadContractPath, makeProvider } from './utils'
import dotenv from 'dotenv'

interface Transmission {
  answer: number
  block_num: number
  observation_timestamp: number
  transmission_timestamp: number
}

const CONTRACT_NAME = 'MockAggregator'
const CONTRACT_PATH = '../../../../contracts/target/release/chainlink_MockAggregator'
let account: Account
let mock: Contract
let transmission: Transmission
let provider: Provider

dotenv.config({ path: __dirname + '/../.env' })

const rl = require('readline').createInterface({
  input: process.stdin,
  output: process.stdout,
})

async function updateLatestRound() {
  provider = makeProvider()

  transmission = {
    answer: 0,
    block_num: 0,
    observation_timestamp: 0,
    transmission_timestamp: 0,
  }

  const privateKey = process.env.DEPLOYER_PRIVATE_KEY as string
  account = new Account(provider, process.env.DEPLOYER_ACCOUNT_ADDRESS as string, privateKey)

  const MockArtifact = loadContractPath(`${CONTRACT_PATH}.sierra`) as CompiledContract

  mock = new Contract(MockArtifact.abi, process.env.MOCK as string)
  mock.connect(account)

  transmission.answer = Number(await input('Enter a number for new answer: '))
  transmission.block_num = Number(await input('Enter a number for new block_num: '))
  transmission.observation_timestamp = Number(
    await input('Enter a number for new observation_timestamp: '),
  )
  transmission.transmission_timestamp = Number(
    await input('Enter a number for new transmission_timestamp: '),
  )
  rl.close()

  await callFunction(transmission)
}

async function callFunction(transmission: Transmission) {
  const tx = await mock.invoke(
    'set_latest_round_data',
    [
      transmission.answer,
      transmission.block_num,
      transmission.observation_timestamp,
      transmission.transmission_timestamp,
    ]
  )

  console.log('Waiting for Tx to be Accepted on Starknet: Updating Latest Round')
  await provider.waitForTransaction(tx.transaction_hash)
}

function input(prompt: string) {
  return new Promise((callbackFn, errorFn) => {
    rl.question(prompt, (uinput: string) => {
      switch (isNaN(Number(uinput))) {
        case true:
          console.log('input is not a number we will use the default value of 1')
          uinput = '1'
          break
        default:
          break
      }
      callbackFn(uinput)
    })
  })
}

updateLatestRound()
