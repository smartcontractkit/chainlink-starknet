import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { HttpNetworkConfig } from 'hardhat/types'

import dotenv from 'dotenv'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import { Account, CompiledContract, Contract as StarknetContract } from 'starknet'
import {
  createDeployerAccount,
  loadContractPath,
  loadContract_Solidity,
  loadContract_Solidity_V8,
  makeProvider,
} from './utils'

dotenv.config({ path: __dirname + '/../.env' })
const UPTIME_FEED_PATH = '../../../../contracts/target/release/chainlink_SequencerUptimeFeed'

// TODO: need to modify when the cairo-1.0 branch is rebased to include StarknetValidator changes

let validator: Contract
let mockStarknetMessengerFactory: ContractFactory
let mockStarknetMessenger: Contract
let deployer: SignerWithAddress
let eoaValidator: SignerWithAddress
let networkUrl: string
let account: Account

let mockGasPriceFeed: MockContract
let mockAccessController: MockContract
let mockAggregator: MockContract

export async function consumerValidator() {
  const provider = makeProvider()

  account = createDeployerAccount(provider)

  networkUrl = (network.config as HttpNetworkConfig).url
  const accounts = await ethers.getSigners()
  deployer = accounts[0]
  eoaValidator = accounts[1]

  const aggregatorAbi = loadContract_Solidity_V8('AggregatorV3Interface')
  const accessControllerAbi = loadContract_Solidity_V8('AccessControllerInterface')

  // Deploy the mock feed
  mockGasPriceFeed = await deployMockContract(deployer, aggregatorAbi.abi)
  await mockGasPriceFeed.mock.latestRoundData.returns(
    '73786976294838220258' /** roundId */,
    '96800000000' /** answer */,
    '163826896' /** startedAt */,
    '1638268960' /** updatedAt */,
    '73786976294838220258' /** answeredInRound */,
  )

  // Deploy the mock access controller
  mockAccessController = await deployMockContract(deployer, accessControllerAbi.abi)

  // Deploy the mock aggregator
  mockAggregator = await deployMockContract(deployer, aggregatorAbi.abi)
  await mockAggregator.mock.latestRoundData.returns(
    '73786976294838220258' /** roundId */,
    1 /** answer */,
    '163826896' /** startedAt */,
    '1638268960' /** updatedAt */,
    '73786976294838220258' /** answeredInRound */,
  )

  const validatorArtifact = await loadContract_Solidity('emergency', 'StarknetValidator')
  const validatorFactory = await ethers.getContractFactoryFromArtifact(validatorArtifact, deployer)

  const mockStarknetMessagingArtifact = await loadContract_Solidity(
    'mocks',
    'MockStarknetMessaging',
  )
  mockStarknetMessengerFactory = await ethers.getContractFactoryFromArtifact(
    mockStarknetMessagingArtifact,
    deployer,
  )

  const messageCancellationDelay = 5 * 60 // seconds
  mockStarknetMessenger = await mockStarknetMessengerFactory.deploy(messageCancellationDelay)
  await mockStarknetMessenger.deployed()

  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH) as CompiledContract

  const mockUptimeFeedDeploy = new StarknetContract(
    UptimeFeedArtifact.abi,
    process.env.UPTIME_FEED as string,
    provider,
  )

  mockUptimeFeedDeploy.connect(account)

  validator = await validatorFactory.deploy(
    mockStarknetMessenger.address,
    mockAccessController.address,
    mockGasPriceFeed.address,
    mockAggregator.address,
    mockUptimeFeedDeploy.address,
    0,
  )

  console.log('Validator address: ', validator.address)

  const tx = await mockUptimeFeedDeploy.invoke('set_l1_sender', [validator.address])

  await provider.waitForTransaction(tx.transaction_hash)

  await validator.addAccess(eoaValidator.address)
  setInterval(callFunction, 60_000)
}

async function callFunction() {
  await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessenger.address)

  await validator.connect(eoaValidator).validate(0, 0, 1, 1)

  const flushL1Response = await starknet.devnet.flush()
  flushL1Response.consumed_messages.from_l1
}
consumerValidator()
