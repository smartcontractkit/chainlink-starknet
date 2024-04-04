import { abi as starknetMessagingAbi } from '../../artifacts/vendor/starkware-libs/cairo-lang/src/starkware/starknet/solidity/IStarknetMessaging.sol/IStarknetMessaging.json'
import { abi as accessControllerAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol/AccessControllerInterface.json'
import { abi as aggregatorAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol/AggregatorV3Interface.json'
import { fetchStarknetAccount, getStarknetContractArtifacts, waitForTransactions } from '../utils'
import { Contract as StarknetContract, RpcProvider, CallData, Account, hash } from 'starknet'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import { BigNumber, Contract as EthersContract, ContractFactory } from 'ethers'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import * as l1l2messaging from '../l1-l2-messaging'
import { STARKNET_DEVNET_URL } from '../constants'
import * as account from '../account'
import { ethers } from 'hardhat'
import { expect } from 'chai'

describe('StarknetValidator', () => {
  const provider = new RpcProvider({ nodeUrl: STARKNET_DEVNET_URL })
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let defaultAccount: Account
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress
  let alice: SignerWithAddress

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

    // Fetch predefined L1 EOA accounts
    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]
    alice = accounts[2]

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

  describe('#constructor', () => {
    it('reverts when the StarknetMessaging address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          ethers.constants.AddressZero,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          l2Contract.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidStarknetMessagingAddress')
    })

    it('reverts when the L2 feed is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarknetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          0,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidL2FeedAddress')
    })

    it('reverts when the Aggregator address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarknetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          ethers.constants.AddressZero,
          l2Contract.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidSourceAggregatorAddress')
    })

    it('reverts when the Access Controller address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarknetMessaging.address,
          ethers.constants.AddressZero,
          mockGasPriceFeed.address,
          mockAggregator.address,
          l2Contract.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidAccessControllerAddress')
    })

    it('reverts when the L1 Gas Price feed address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarknetMessaging.address,
          mockAccessController.address,
          ethers.constants.AddressZero,
          mockAggregator.address,
          l2Contract.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
    })

    it('is initialized with the correct gas config', async () => {
      const gasConfig = await starknetValidator.getGasConfig()
      expect(gasConfig.gasEstimate).to.equal(0) // Initialized with 0 in before function
      expect(gasConfig.gasPriceL1Feed).to.hexEqual(mockGasPriceFeed.address)
      expect(gasConfig.gasAdjustment).to.equal(0)
    })

    it('is initialized with the correct access controller address', async () => {
      const acAddr = await starknetValidator.getConfigAC()
      expect(acAddr).to.hexEqual(mockAccessController.address)
    })

    it('is initialized with the correct source aggregator address', async () => {
      const aggregatorAddr = await starknetValidator.getSourceAggregator()
      expect(aggregatorAddr).to.hexEqual(mockAggregator.address)
    })

    it('should get the selector from the name successfully', async () => {
      const actual = hash.getSelectorFromName('update_status')
      const expected = 1585322027166395525705364165097050997465692350398750944680096081848180365267n
      expect(BigInt(actual)).to.equal(expected)

      const computedActual = await starknetValidator.SELECTOR_STARK_UPDATE_STATUS()
      expect(computedActual).to.equal(expected)
    })
  })

  describe('#retry', () => {
    describe('when called by account with no access', () => {
      it('reverts', async () => {
        await expect(starknetValidator.connect(alice).retry()).to.be.revertedWith('No access')
      })
    })

    describe('when called by account with access', () => {
      it('transaction succeeds', async () => {
        const waffleMockStarknetMessaging = await deployMockContract(deployer, starknetMessagingAbi)
        await waffleMockStarknetMessaging.mock.sendMessageToL2.returns(
          ethers.utils.formatBytes32String('0'),
          0,
        )
        await mockAggregator.mock.latestRoundData.returns(
          '0' /** roundId */,
          1 /** answer */,
          '0' /** startedAt */,
          '0' /** updatedAt */,
          '0' /** answeredInRound */,
        )
        await mockGasPriceFeed.mock.latestRoundData.returns(
          '0' /** roundId */,
          1 /** answer */,
          '0' /** startedAt */,
          '0' /** updatedAt */,
          '0' /** answeredInRound */,
        )

        const starknetValidator = await starknetValidatorFactory.deploy(
          waffleMockStarknetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          l2Contract.address,
          0,
          0,
        )

        await starknetValidator.addAccess(deployer.address)

        await starknetValidator.retry()
      })
    })
  })

  describe('#setConfigAC', () => {
    describe('when called by non owner', () => {
      it('reverts', async () => {
        const newACAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starknetValidator.connect(alice).setConfigAC(newACAddr)).to.be.revertedWith(
          'Only callable by owner',
        )
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        const newACAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starknetValidator.setConfigAC(newACAddr))
          .to.emit(starknetValidator, 'ConfigACSet')
          .withArgs(mockAccessController.address, newACAddr)
      })

      it('sets the access controller address', async () => {
        const newACAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await starknetValidator.connect(deployer).setConfigAC(newACAddr)
        expect(await starknetValidator.getConfigAC()).to.equal(newACAddr)
      })

      it('no-op if new address equals previous address', async () => {
        const tx = await starknetValidator.setConfigAC(mockAccessController.address)
        const receipt = await tx.wait()
        expect(receipt.events).is.empty
        expect(await starknetValidator.getConfigAC()).to.equal(mockAccessController.address)
      })

      it('reverts if address is zero', async () => {
        await expect(
          starknetValidator.connect(deployer).setConfigAC(ethers.constants.AddressZero),
        ).to.be.revertedWithCustomError(starknetValidator, 'InvalidAccessControllerAddress')
      })
    })
  })

  describe('#setSourceAggregator', () => {
    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starknetValidator.connect(alice).setSourceAggregator(ethers.constants.AddressZero),
        ).to.be.revertedWith('Only callable by owner')
      })
    })

    describe('when source address is the zero address', () => {
      it('reverts', async () => {
        await expect(
          starknetValidator.connect(deployer).setSourceAggregator(ethers.constants.AddressZero),
        ).to.be.revertedWithCustomError(starknetValidator, 'InvalidSourceAggregatorAddress')
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        const newSourceAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starknetValidator.setSourceAggregator(newSourceAddr))
          .to.emit(starknetValidator, 'SourceAggregatorSet')
          .withArgs(mockAggregator.address, newSourceAddr)
      })

      it('sets the source aggregator address', async () => {
        expect(await starknetValidator.getSourceAggregator()).to.hexEqual(mockAggregator.address)
      })

      it('does nothing if new address equal to previous', async () => {
        const tx = await starknetValidator
          .connect(deployer)
          .setSourceAggregator(mockAggregator.address)
        const receipt = await tx.wait()
        expect(receipt.events).to.be.empty

        expect(await starknetValidator.getSourceAggregator()).to.hexEqual(mockAggregator.address)
      })
    })
  })

  describe('#setGasConfig', () => {
    describe('when called by non owner without access', () => {
      beforeEach(async () => {
        await mockAccessController.mock.hasAccess.returns(false)
      })

      it('reverts', async () => {
        await expect(
          starknetValidator.connect(alice).setGasConfig(0, mockGasPriceFeed.address, 0),
        ).to.be.revertedWithCustomError(starknetValidator, 'AccessForbidden')
      })
    })

    describe('when called by owner', () => {
      it('correctly sets the gas config', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        // gasAdjustment of 110 equates to 1.1x
        const newGasAdjustment = 110
        await starknetValidator
          .connect(deployer)
          .setGasConfig(newGasEstimate, newFeedAddr, newGasAdjustment)
        const gasConfig = await starknetValidator.getGasConfig()
        expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
        expect(gasConfig.gasPriceL1Feed).to.hexEqual(newFeedAddr)
        expect(gasConfig.gasAdjustment).to.equal(newGasAdjustment)
      })

      it('emits an event', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        const newGasAdjustment = 110
        await expect(
          starknetValidator
            .connect(deployer)
            .setGasConfig(newGasEstimate, newFeedAddr, newGasAdjustment),
        )
          .to.emit(starknetValidator, 'GasConfigSet')
          .withArgs(newGasEstimate, newFeedAddr, newGasAdjustment)
      })

      describe('when l1 gas price feed address is the zero address', () => {
        it('reverts', async () => {
          await expect(
            starknetValidator
              .connect(deployer)
              .setGasConfig(25000, ethers.constants.AddressZero, 0),
          ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
        })
      })
    })

    describe('when called by an address with access', () => {
      beforeEach(async () => {
        await mockAccessController.mock.hasAccess.returns(true)
      })

      it('correctly sets the gas config', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        const newGasAdjustment = 110
        await starknetValidator
          .connect(eoaValidator)
          .setGasConfig(newGasEstimate, newFeedAddr, newGasAdjustment)
        const gasConfig = await starknetValidator.getGasConfig()
        expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
        expect(gasConfig.gasPriceL1Feed).to.hexEqual(newFeedAddr)
        expect(gasConfig.gasAdjustment).to.equal(newGasAdjustment)
      })

      it('emits an event', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        const newGasAdjustment = 110
        await expect(
          starknetValidator
            .connect(eoaValidator)
            .setGasConfig(newGasEstimate, newFeedAddr, newGasAdjustment),
        )
          .to.emit(starknetValidator, 'GasConfigSet')
          .withArgs(newGasEstimate, newFeedAddr, newGasAdjustment)
      })

      describe('when l1 gas price feed address is the zero address', () => {
        it('reverts', async () => {
          await expect(
            starknetValidator
              .connect(eoaValidator)
              .setGasConfig(25000, ethers.constants.AddressZero, 0),
          ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
        })
      })
    })
  })

  describe('#approximateGasPrice', () => {
    it('calculates gas price with scalar coefficient', async () => {
      await mockGasPriceFeed.mock.latestRoundData.returns(
        '0' /** roundId */,
        96800000000 /** answer */,
        '0' /** startedAt */,
        '0' /** updatedAt */,
        '0' /** answeredInRound */,
      )
      // 96800000000 is the mocked value from gas feed
      const expectedGasPrice = BigNumber.from(96800000000).mul(110).div(100)

      await starknetValidator.connect(deployer).setGasConfig(0, mockGasPriceFeed.address, 110)

      const gasPrice = await starknetValidator.connect(deployer).approximateGasPrice()

      expect(gasPrice).to.equal(expectedGasPrice)
    })
  })

  describe('#validate', () => {
    beforeEach(async () => {
      await expect(
        deployer.sendTransaction({ to: starknetValidator.address, value: 100n }),
      ).to.changeEtherBalance(starknetValidator, 100n)
    })

    it('reverts if `StarknetValidator.validate` called by account with no access', async () => {
      const c = starknetValidator.connect(eoaValidator)
      await expect(c.validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should not revert if `sequencer_uptime_feed.latest_round_data` called by an Account with no explicit access (Accounts are allowed read access)', async () => {
      const result = await l2Contract.latest_round_data()
      expect(result['answer']).to.equal('0')
    })

    it('should deploy the messaging contract', async () => {
      const { messaging_contract_address } = await l1l2messaging.loadL1MessagingContract({
        address: mockStarknetMessaging.address,
      })
      expect(messaging_contract_address).not.to.be.undefined
    })

    it('should load the already deployed contract if the address is provided', async () => {
      const { messaging_contract_address } = await l1l2messaging.loadL1MessagingContract({
        address: mockStarknetMessaging.address,
      })
      expect(mockStarknetMessaging.address).to.hexEqual(messaging_contract_address)
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

      // Logging (to help debug potential flaky test)
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

      // Logging (to help debug potential flaky test)
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
      const messages = new Array<l1l2messaging.FlushedMessages>()
      const newGasEstimate = 1
      const receipts = await waitForTransactions(
        [
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
        ],
        async () => {
          // Simulate the L1 - L2 comms
          const resp = await l1l2messaging.flush()
          if (resp.messages_to_l2.length !== 0) {
            expect(resp.messages_to_l1).to.be.empty

            const msgFromL1 = resp.messages_to_l2
            expect(msgFromL1).to.have.a.lengthOf(1)
            expect(msgFromL1[0].l1_contract_address).to.hexEqual(starknetValidator.address)
            expect(msgFromL1[0].l2_contract_address).to.hexEqual(l2Contract.address)

            messages.push(resp)
          }
        },
      )

      // Makes sure the correct number of messages were transmitted
      expect(messages.length).to.eq(4)

      // Assert L2 effects
      const result = await l2Contract.latest_round_data()

      // Logging (to help debug potential flaky test)
      console.log(
        JSON.stringify(
          {
            latestRoundData: result,
            flushResponse: messages,
            txReceipts: receipts,
          },
          (_, value) => (typeof value === 'bigint' ? value.toString() : value),
          2,
        ),
      )

      expect(result['answer']).to.equal('0') // final status 0
    })
  })

  describe('#withdrawFunds', () => {
    beforeEach(async () => {
      await expect(() =>
        deployer.sendTransaction({ to: starknetValidator.address, value: 10 }),
      ).to.changeEtherBalance(starknetValidator, 10n)
    })

    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(starknetValidator.connect(alice).withdrawFunds()).to.be.revertedWith(
          'Only callable by owner',
        )
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        await expect(starknetValidator.connect(deployer).withdrawFunds())
          .to.emit(starknetValidator, 'FundsWithdrawn')
          .withArgs(deployer.address, 10)
      })

      it('withdraws all funds to deployer', async () => {
        await starknetValidator.connect(deployer).withdrawFunds()
        const balance = await ethers.provider.getBalance(starknetValidator.address)
        expect(balance).to.equal(0n)
      })
    })
  })

  describe('#withdrawFundsTo', () => {
    beforeEach(async () => {
      await expect(() =>
        deployer.sendTransaction({ to: starknetValidator.address, value: 10 }),
      ).to.changeEtherBalance(starknetValidator, 10)
    })

    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starknetValidator.connect(alice).withdrawFundsTo(alice.address),
        ).to.be.revertedWith('Only callable by owner')
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        await expect(starknetValidator.connect(deployer).withdrawFundsTo(eoaValidator.address))
          .to.emit(starknetValidator, 'FundsWithdrawn')
          .withArgs(eoaValidator.address, 10)
      })

      it('withdraws all funds to deployer', async () => {
        await starknetValidator.connect(deployer).withdrawFunds()
        const balance = await ethers.provider.getBalance(starknetValidator.address)
        expect(balance).to.equal(0n)
      })
    })
  })
})
