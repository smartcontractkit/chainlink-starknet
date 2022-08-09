import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { HttpNetworkConfig, StarknetContract, Account } from 'hardhat/types'
import fs from 'fs'
import dotenv from 'dotenv'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { defaultProvider, ec } from 'starknet'
import { loadContractSequencer } from '.'

dotenv.config({ path: __dirname + '/.env' })
const AGGREGATOR_NAME = 'Mock_Aggregator'
const UPTIME_FEED_NAME = 'sequencer_uptime_feed'

let Validator: Contract
let MockStarknetMessaging: ContractFactory
let mockStarknetMessaging: Contract
let deployer: SignerWithAddress
let eoaValidator: SignerWithAddress
let networkUrl: string
let account: Account

async function main() {
  // account = await starknet.getAccountFromAddress(
  //   process.env.ACCOUNT_ADDRESS as string,
  //   process.env.PRIVATE_KEY as string,
  //   'OpenZeppelin',
  // )

  
  // const MockUptimeFeedFactory = loadContractSequencer(UPTIME_FEED_NAME)
  const MockUptimeFeedFactory = await starknet.getContractFactory('contracts/contracts/cairo/ocr2/SequencerUptimeFeed/sequencer_uptime_feed.cairo')
  const MockUptimeFeedDeploy = await MockUptimeFeedFactory.deploy({})
  // const MockUptimeFeedDeploy = await defaultProvider.deployContract({
  //   contract: MockUptimeFeedFactory,
  //   constructorCalldata: [0, account.address],
  // })

  const AggregatorFactory = await starknet.getContractFactory(AGGREGATOR_NAME)
  const AggregatorDeploy = await AggregatorFactory.deploy({})

  fs.appendFile(__dirname + '/.env', '\nUPTIME_FEED=' + MockUptimeFeedDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nMOCK_AGGREGATOR=' + AggregatorDeploy.address, function (err) {
    if (err) throw err
  })

  networkUrl = (network.config as HttpNetworkConfig).url
  const accounts = await ethers.getSigners()
  deployer = accounts[0]
  eoaValidator = accounts[1]

  const ValidatorFactory = await ethers.getContractFactory('Validator', deployer)
  MockStarknetMessaging = await ethers.getContractFactory('MockStarknetMessaging', deployer)

  mockStarknetMessaging = await MockStarknetMessaging.deploy()
  await mockStarknetMessaging.deployed()

  Validator = await ValidatorFactory.deploy(mockStarknetMessaging.address, MockUptimeFeedDeploy.address)
  console.log('Validator address: ', Validator.address)

  // await account.invoke(MockUptimeFeedDeploy, 'set_l1_sender', { address: Validator.address })

  // await Validator.addAccess(eoaValidator.address)
  // setInterval(callFunction, 60_000)
}

async function callFunction() {
  await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

  await Validator.connect(eoaValidator).validate(0, 0, 1, 0)

  const flushL1Response = await starknet.devnet.flush()
  flushL1Response.consumed_messages.from_l1
}
main()
