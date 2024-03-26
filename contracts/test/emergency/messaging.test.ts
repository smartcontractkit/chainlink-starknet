import { abi as accessControllerAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol/AccessControllerInterface.json'
import { abi as aggregatorAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol/AggregatorV3Interface.json'
import { fetchStarknetAccount, getStarknetContractArtifacts, waitForTransactions } from '../utils'
import { Contract as StarknetContract, RpcProvider, CallData, Account } from 'starknet'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { Contract as EthersContract, ContractFactory } from 'ethers'
import * as l1l2messaging from '../l1-l2-messaging'
import { STARKNET_DEVNET_URL } from '../constants'
import * as account from '../account'
import { ethers } from 'hardhat'
import { expect } from 'chai'

describe('StarknetMessaging', () => {
  const provider = new RpcProvider({ nodeUrl: STARKNET_DEVNET_URL })
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let defaultAccount: Account
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let starknetValidatorFactory: ContractFactory
  let starknetValidator: EthersContract
  let mockStarknetMessagingFactory: ContractFactory
  let mockStarknetMessaging: EthersContract
  let mockGasPriceFeed: MockContract
  let mockAccessController: MockContract
  let mockAggregator: MockContract

  let l2Contract: StarknetContract

  before(async () => {
    // Setup L2 account
    defaultAccount = await fetchStarknetAccount()
    await funder.fund([{ account: defaultAccount.address, amount: 1e21 }])

    // Fetch predefined L1 EOA accounts
    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]

    // Deploy the mock feed
    mockGasPriceFeed = await deployMockContract(deployer, aggregatorAbi)
    await mockGasPriceFeed.mock.latestRoundData.returns(
      '73786976294838220258' /** roundId */,
      '96800000000' /** answer */,
      '163826896' /** startedAt */,
      '1638268960' /** updatedAt */,
      '73786976294838220258' /** answeredInRound */,
    )

    // Deploy the mock access controller
    mockAccessController = await deployMockContract(deployer, accessControllerAbi)

    // Deploy the mock aggregator
    mockAggregator = await deployMockContract(deployer, aggregatorAbi)
    await mockAggregator.mock.latestRoundData.returns(
      '73786976294838220258' /** roundId */,
      1 /** answer */,
      '163826896' /** startedAt */,
      '1638268960' /** updatedAt */,
      '73786976294838220258' /** answeredInRound */,
    )
  })

  beforeEach(async () => {
    // Deploy the MockStarknetMessaging contract used to simulate L1 - L2 comms
    mockStarknetMessagingFactory = await ethers.getContractFactory(
      'MockStarknetMessaging',
      deployer,
    )
    const messageCancellationDelay = 5 * 60 // seconds
    mockStarknetMessaging = await mockStarknetMessagingFactory.deploy(messageCancellationDelay)
    await mockStarknetMessaging.deployed()

    // Deploy L2 feed contract
    const ddL2Contract = await defaultAccount.declareAndDeploy({
      ...getStarknetContractArtifacts('SequencerUptimeFeed'),
      constructorCalldata: CallData.compile({
        initial_status: 0,
        owner_address: defaultAccount.address,
      }),
    })

    // Creates a starknet contract instance for the l2 feed
    const { abi: l2FeedAbi } = await provider.getClassByHash(ddL2Contract.declare.class_hash)
    l2Contract = new StarknetContract(l2FeedAbi, ddL2Contract.deploy.address, provider)

    // Deploy the L1 StarknetValidator
    starknetValidatorFactory = await ethers.getContractFactory('StarknetValidator', deployer)
    starknetValidator = await starknetValidatorFactory.deploy(
      mockStarknetMessaging.address,
      mockAccessController.address,
      mockGasPriceFeed.address,
      mockAggregator.address,
      l2Contract.address,
      0,
      0,
    )

    // Point the L2 feed contract to receive from the L1 StarknetValidator contract
    await defaultAccount.execute(
      l2Contract.populate('set_l1_sender', {
        address: starknetValidator.address,
      }),
    )
  })
  describe('#validate', () => {
    beforeEach(async () => {
      await expect(
        deployer.sendTransaction({ to: starknetValidator.address, value: 100n }),
      ).to.changeEtherBalance(starknetValidator, 100n)
    })

    it('should send a message to the L2 contract', async () => {
      // Load the mock messaging contract
      await l1l2messaging.loadL1MessagingContract({ address: mockStarknetMessaging.address })

      // Return gas price of 1
      await mockGasPriceFeed.mock.latestRoundData.returns(
        '0' /** roundId */,
        1 /** answer */,
        '0' /** startedAt */,
        '0' /** updatedAt */,
        '0' /** answeredInRound */,
      )

      // Simulate L1 transmit + validate
      const newGasEstimate = 1
      const receipts = await waitForTransactions([
        // Add access
        () => starknetValidator.addAccess(eoaValidator.address),

        // By default the gas config is 0, we need to change it or we will submit a 0 fee
        () =>
          starknetValidator
            .connect(deployer)
            .setGasConfig(newGasEstimate, mockGasPriceFeed.address, 100),

        // gasPrice (1) * newGasEstimate (1)
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 1),
      ])

      // Simulate the L1 - L2 comms
      const resp = await l1l2messaging.flush()
      const msgFromL1 = resp.messages_to_l2
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.messages_to_l1).to.be.empty

      expect(msgFromL1.at(0)?.l1_contract_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1.at(0)?.l2_contract_address).to.hexEqual(l2Contract.address)

      // Assert L2 effects
      const result = await l2Contract.latest_round_data()

      // TODO: remove once flaky test is fixed
      // Logging (to help debug flaky test)
      console.log(
        JSON.stringify(
          {
            latestRoundData: result,
            flushResponse: resp,
            txReceipts: receipts,
          },
          (_, value) => (typeof value === 'bigint' ? value.toString() : value),
          2,
        ),
      )

      expect(result['answer']).to.equal('1')
    })

    it('should always send a **boolean** message to L2 contract', async () => {
      // Load the mock messaging contract
      await l1l2messaging.loadL1MessagingContract({ address: mockStarknetMessaging.address })

      // Return gas price of 1
      await mockGasPriceFeed.mock.latestRoundData.returns(
        '0' /** roundId */,
        1 /** answer */,
        '0' /** startedAt */,
        '0' /** updatedAt */,
        '0' /** answeredInRound */,
      )

      // Simulate L1 transmit + validate
      const newGasEstimate = 1
      const receipts = await waitForTransactions([
        // Add access
        () => starknetValidator.connect(deployer).addAccess(eoaValidator.address),

        // By default the gas config is 0, we need to change it or we will submit a 0 fee
        () =>
          starknetValidator
            .connect(deployer)
            .setGasConfig(newGasEstimate, mockGasPriceFeed.address, 100),

        // Incorrect value
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 127),
      ])

      // Simulate the L1 - L2 comms
      const resp = await l1l2messaging.flush()
      const msgFromL1 = resp.messages_to_l2
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.messages_to_l1).to.be.empty

      expect(msgFromL1[0].l1_contract_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1[0].l2_contract_address).to.hexEqual(l2Contract.address)

      // Assert L2 effects
      const result = await l2Contract.latest_round_data()

      // TODO: remove once flaky test is fixed
      // Logging (to help debug flaky test)
      console.log(
        JSON.stringify(
          {
            latestRoundData: result,
            flushResponse: resp,
            txReceipts: receipts,
          },
          (_, value) => (typeof value === 'bigint' ? value.toString() : value),
          2,
        ),
      )

      expect(result['answer']).to.equal('0') // status unchanged - incorrect value treated as false
    })

    it('should send multiple messages', async () => {
      // Load the mock messaging contract
      await l1l2messaging.loadL1MessagingContract({ address: mockStarknetMessaging.address })

      // Return gas price of 1
      await mockGasPriceFeed.mock.latestRoundData.returns(
        '0' /** roundId */,
        1 /** answer */,
        '0' /** startedAt */,
        '0' /** updatedAt */,
        '0' /** answeredInRound */,
      )

      // Simulate L1 transmit + validate
      const newGasEstimate = 1
      const receipts = await waitForTransactions([
        // Add access
        () => starknetValidator.connect(deployer).addAccess(eoaValidator.address),

        // By default the gas config is 0, we need to change it or we will submit a 0 fee
        () =>
          starknetValidator
            .connect(deployer)
            .setGasConfig(newGasEstimate, mockGasPriceFeed.address, 100),

        // Validate
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 1),
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 1),
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 127), // incorrect value
        () => starknetValidator.connect(eoaValidator).validate(0, 0, 1, 0), // final status
      ])

      // Simulate the L1 - L2 comms
      const resp = await l1l2messaging.flush()
      const msgFromL1 = resp.messages_to_l2
      expect(msgFromL1).to.have.a.lengthOf(4)
      expect(resp.messages_to_l1).to.be.empty

      expect(msgFromL1[0].l1_contract_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1[0].l2_contract_address).to.hexEqual(l2Contract.address)

      // Assert L2 effects
      const result = await l2Contract.latest_round_data()

      // TODO: remove once flaky test is fixed
      // Logging (to help debug flaky test)
      console.log(
        JSON.stringify(
          {
            latestRoundData: result,
            flushResponse: resp,
            txReceipts: receipts,
          },
          (_, value) => (typeof value === 'bigint' ? value.toString() : value),
          2,
        ),
      )

      expect(result['answer']).to.equal('0') // final status 0
    })
  })
})
