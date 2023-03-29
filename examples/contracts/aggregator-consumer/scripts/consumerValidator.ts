import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { HttpNetworkConfig } from 'hardhat/types'

import dotenv from 'dotenv'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import { Account, Contract as StarknetContract } from 'starknet'
import {
  createDeployerAccount,
  loadContractPath,
  loadContract_Solidity,
  loadContract_Solidity_V8,
  makeProvider,
} from './utils'

dotenv.config({ path: __dirname + '/../.env' })
const UPTIME_FEED_PATH =
  '../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed'
const UPTIME_FEED_NAME = 'sequencer_uptime_feed'

let validator: Contract
let mockStarkNetMessengerFactory: ContractFactory
let mockStarkNetMessenger: Contract
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

  const validatorArtifact = await loadContract_Solidity('emergency', 'StarkNetValidator')
  const validatorFactory = await ethers.getContractFactoryFromArtifact(validatorArtifact, deployer)

  const mockStarknetMessagingArtifact = await loadContract_Solidity(
    'mocks',
    'MockStarkNetMessaging',
  )
  mockStarkNetMessengerFactory = await ethers.getContractFactoryFromArtifact(
    mockStarknetMessagingArtifact,
    deployer,
  )

  const messageCancellationDelay = 5 * 60 // seconds
  mockStarkNetMessenger = await mockStarkNetMessengerFactory.deploy(messageCancellationDelay)
  await mockStarkNetMessenger.deployed()

  const UptimeFeedArtifact = loadContractPath(UPTIME_FEED_PATH, UPTIME_FEED_NAME)

  const mockUptimeFeedDeploy = new StarknetContract(
    UptimeFeedArtifact.abi,
    process.env.UPTIME_FEED as string,
    provider,
  )

  validator = await validatorFactory.deploy(
    mockStarkNetMessenger.address,
    mockAccessController.address,
    mockGasPriceFeed.address,
    mockAggregator.address,
    mockUptimeFeedDeploy.address,
    0,
  )

  console.log('Validator address: ', validator.address)
  const transaction = await account.execute(
    {
      contractAddress: mockUptimeFeedDeploy.address,
      entrypoint: 'set_l1_sender',
      calldata: [validator.address],
    },
    [UptimeFeedArtifact.abi],
  )

  await provider.waitForTransaction(transaction.transaction_hash)

  await validator.addAccess(eoaValidator.address)
  setInterval(callFunction, 60_000)
}

async function callFunction() {
  await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessenger.address)

  await validator.connect(eoaValidator).validate(0, 0, 1, 1)

  const flushL1Response = await starknet.devnet.flush()
  flushL1Response.consumed_messages.from_l1
}
consumerValidator()
