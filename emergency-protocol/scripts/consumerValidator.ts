import { ethers, starknet, network } from 'hardhat'
import { BigNumber, Contract, ContractFactory } from 'ethers'
import { HttpNetworkConfig, StarknetContract } from 'hardhat/types'
import { expect } from 'chai'
import { loadContract } from './index'
import { defaultProvider, ec, Account } from 'starknet'
import fs from 'fs'
import dotenv from 'dotenv'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
/// Pick ABIs from compilation
// @ts-ignore
import { abi as optimismSequencerStatusRecorderAbi } from '../../../artifacts/src/v0.8/dev/OptimismSequencerUptimeFeed.sol/OptimismSequencerUptimeFeed.json'
// @ts-ignore
import { abi as optimismL1CrossDomainMessengerAbi } from '@eth-optimism/contracts/artifacts/contracts/L1/messaging/L1CrossDomainMessenger.sol'
// @ts-ignore
import { abi as aggregatorAbi } from '../../../artifacts/src/v0.8/interfaces/AggregatorV2V3Interface.sol/AggregatorV2V3Interface.json'

dotenv.config({ path: __dirname + '/.env' })
const CONSUMER_NAME = 'Mock_Aggregator'
const MOCK_NAME = 'Mock_Uptime_feed'
const PRICE_CONSUMER_NAME = 'Price_Consumer'
const DECIMALS = 18

function adaptAddress(address: string) {
  return '0x' + BigInt(address).toString(16)
}


export function expectAddressEquality(actual: string, expected: string) {
  expect(adaptAddress(actual)).to.equal(adaptAddress(expected))
}

let Validator: Contract
let MockStarknetMessaging: ContractFactory
let mockStarknetMessaging: Contract
let deployer: SignerWithAddress
let eoaValidator: SignerWithAddress
let networkUrl: string

async function main() {
  // const L1_STARKNET_CORE = "0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4"
  networkUrl = (network.config as HttpNetworkConfig).url
  const accounts = await ethers.getSigners()
  deployer = accounts[0]
  eoaValidator = accounts[1]

  const ValidatorFactory = await ethers.getContractFactory('MockValidator', deployer)
  MockStarknetMessaging = await ethers.getContractFactory('MockStarknetMessaging', deployer)

  mockStarknetMessaging = await MockStarknetMessaging.deploy()
  await mockStarknetMessaging.deployed()

  Validator = await ValidatorFactory.deploy(mockStarknetMessaging.address)
  console.log('Validator address: ', Validator.address)

  const MockUptimeFeedFactory = await starknet.getContractFactory(MOCK_NAME)
  const MockUptimeFeedDeploy = await MockUptimeFeedFactory.deploy({ l1_validator_address: Validator.address })

  const AggregatorFactory = await starknet.getContractFactory(CONSUMER_NAME)
  const AggregatorDeploy = await AggregatorFactory.deploy({})

  fs.appendFile(__dirname + '/.env', '\nUPTIME_FEED=' + MockUptimeFeedDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nMOCK_AGGREGATOR=' + AggregatorDeploy.address, function (err) {
    if (err) throw err
  })
  
  Validator.setL2UptimeFeedAdd(MockUptimeFeedDeploy.address)
  // Validator.addAccess(eoaValidator.address)
  // Validator.connect(eoaValidator).validate(0, 0, 1, 1)
  setInterval(callFunction, 60_000)
}

async function callFunction() {

  // Validator.addAccess(eoaValidator.address)

  console.log("Network: ", network)
  await starknet.devnet.loadL1MessagingContract(
    networkUrl,
    mockStarknetMessaging.address,
  )

  await Validator.connect(eoaValidator).validate(0, 0, 1, 1)

  const flushL1Response = await starknet.devnet.flush()
  flushL1Response.consumed_messages.from_l1

}
main()
