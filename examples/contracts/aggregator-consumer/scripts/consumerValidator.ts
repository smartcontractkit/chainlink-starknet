import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { HttpNetworkConfig, StarknetContract, Account } from 'hardhat/types'
import fs from 'fs'
import dotenv from 'dotenv'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'

dotenv.config({ path: __dirname + '/.env' })
const AGGREGATOR_NAME = 'MockAggregator'
const UPTIME_FEED_NAME =
  '../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed/sequencer_uptime_feed'

let validator: Contract
let mockStarkNetMessengerFactory: ContractFactory
let mockStarkNetMessenger: Contract
let deployer: SignerWithAddress
let eoaValidator: SignerWithAddress
let networkUrl: string
let account: Account

async function main() {
  account = await starknet.getAccountFromAddress(
    process.env.ACCOUNT_ADDRESS as string,
    process.env.PRIVATE_KEY as string,
    'OpenZeppelin',
  )

  const mockUptimeFeedFactory = await starknet.getContractFactory(UPTIME_FEED_NAME)
  const mockUptimeFeedDeploy = await mockUptimeFeedFactory.deploy({
    initial_status: 0,
    owner_address: account.starknetContract.address,
  })

  const aggregatorFactory = await starknet.getContractFactory(AGGREGATOR_NAME)
  const aggregatorDeploy = await aggregatorFactory.deploy({})

  fs.appendFile(__dirname + '/.env', '\nUPTIME_FEED=' + mockUptimeFeedDeploy.address, function (
    err,
  ) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nMOCK_AGGREGATOR=' + aggregatorDeploy.address, function (
    err,
  ) {
    if (err) throw err
  })

  networkUrl = (network.config as HttpNetworkConfig).url
  const accounts = await ethers.getSigners()
  deployer = accounts[0]
  eoaValidator = accounts[1]

  const validatorFactory = await ethers.getContractFactory('StarkNetValidator', deployer)
  mockStarkNetMessengerFactory = await ethers.getContractFactory('MockStarkNetMessaging', deployer)

  mockStarkNetMessenger = await mockStarkNetMessengerFactory.deploy()
  await mockStarkNetMessenger.deployed()

  validator = await validatorFactory.deploy(
    mockStarkNetMessenger.address,
    mockUptimeFeedDeploy.address,
  )
  console.log('Validator address: ', validator.address)

  await account.invoke(mockUptimeFeedDeploy, 'set_l1_sender', {
    address: validator.address,
  })

  await validator.addAccess(eoaValidator.address)
  setInterval(callFunction, 60_000)
}

async function callFunction() {
  await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessenger.address)

  await validator.connect(eoaValidator).validate(0, 0, 1, 0)

  const flushL1Response = await starknet.devnet.flush()
  flushL1Response.consumed_messages.from_l1
}
main()
