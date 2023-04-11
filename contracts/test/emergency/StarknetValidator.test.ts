import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { hash, number } from 'starknet'
import {
  Account,
  StarknetContractFactory,
  StarknetContract,
  HttpNetworkConfig,
} from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { abi as aggregatorAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol/AggregatorV3Interface.json'
import { abi as accessControllerAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol/AccessControllerInterface.json'
import { abi as starknetMessagingAbi } from '../../artifacts/vendor/starkware-libs/starkgate-contracts-solidity-v0.8/src/starkware/starknet/solidity/IStarknetMessaging.sol/IStarknetMessaging.json'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import { account, addCompilationToNetwork } from '@chainlink/starknet'
import { StarknetValidator__factory, StarknetValidator } from '../../typechain-types'

describe('StarknetValidator', () => {
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let defaultAccount: Account
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress
  let alice: SignerWithAddress

  let starknetValidatorFactory: StarknetValidator__factory
  let starknetValidator: StarknetValidator
  let mockStarknetMessagingFactory: ContractFactory
  let mockStarknetMessaging: Contract
  let mockGasPriceFeed: MockContract
  let mockAccessController: MockContract
  let mockAggregator: MockContract

  let l2ContractFactory: StarknetContractFactory
  let l2Contract: StarknetContract

  before(async () => {
    await addCompilationToNetwork(
      'src/chainlink/solidity/emergency/StarknetValidator.sol:StarknetValidator',
    )

    // Deploy L2 account
    defaultAccount = await starknet.OpenZeppelinAccount.createAccount()

    // Fund L2 account
    await funder.fund([{ account: defaultAccount.address, amount: 1e21 }])
    await defaultAccount.deployAccount()

    // Fetch predefined L1 EOA accounts
    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]
    alice = accounts[2]

    // Deploy L2 feed contract
    l2ContractFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    await defaultAccount.declare(l2ContractFactory)

    l2Contract = await defaultAccount.deploy(l2ContractFactory, {
      initial_status: 0,
      owner_address: number.toBN(defaultAccount.starknetContract.address),
    })

    // Deploy the MockStarknetMessaging contract used to simulate L1 - L2 comms
    mockStarknetMessagingFactory = await ethers.getContractFactory(
      'MockStarknetMessaging',
      deployer,
    )
    const messageCancellationDelay = 5 * 60 // seconds
    mockStarknetMessaging = await mockStarknetMessagingFactory.deploy(messageCancellationDelay)
    await mockStarknetMessaging.deployed()

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
    // Deploy the L1 StarknetValidator
    const starknetValidatorFactory = await ethers.getContractFactory('StarknetValidator', deployer)
    starknetValidator = await starknetValidatorFactory.deploy(
      mockStarknetMessaging.address,
      mockAccessController.address,
      mockGasPriceFeed.address,
      mockAggregator.address,
      l2Contract.address,
      0,
    )

    // Point the L2 feed contract to receive from the L1 StarknetValidator contract
    await defaultAccount.invoke(l2Contract, 'set_l1_sender', { address: starknetValidator.address })
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
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidStarknetMessagingAddress')
    })

    it('reverts when the L2 feed is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidL2FeedAddress')
    })

    it('reverts when the Aggregator address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          ethers.constants.AddressZero,
          l2Contract.address,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidSourceAggregatorAddress')
    })

    it('reverts when the L1 Gas Price feed address is zero', async () => {
      await expect(
        starknetValidatorFactory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          ethers.constants.AddressZero,
          mockAggregator.address,
          l2Contract.address,
          0,
        ),
      ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
    })

    it('is initialized with the correct gas config', async () => {
      const gasConfig = await starknetValidator.getGasConfig()
      expect(gasConfig.gasEstimate).to.equal(0) // Initialized with 0 in before function
      expect(gasConfig.gasPriceL1Feed).to.hexEqual(mockGasPriceFeed.address)
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
        const waffleMockStarkNetMessaging = await deployMockContract(deployer, starknetMessagingAbi)
        await waffleMockStarkNetMessaging.mock.sendMessageToL2.returns(
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
          waffleMockStarkNetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          l2Contract.address,
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
        await expect(
          starknetValidator.connect(alice).setConfigAC(ethers.constants.AddressZero),
        ).to.be.revertedWith('Only callable by owner')
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
          starknetValidator.connect(alice).setGasConfig(0, mockGasPriceFeed.address),
        ).to.be.revertedWithCustomError(starknetValidator, 'AccessForbidden')
      })
    })

    describe('when called by owner', () => {
      it('correctly sets the gas config', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await starknetValidator.connect(deployer).setGasConfig(newGasEstimate, newFeedAddr)
        const gasConfig = await starknetValidator.getGasConfig()
        expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
        expect(gasConfig.gasPriceL1Feed).to.hexEqual(newFeedAddr)
      })

      it('emits an event', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starknetValidator.connect(deployer).setGasConfig(newGasEstimate, newFeedAddr))
          .to.emit(starknetValidator, 'GasConfigSet')
          .withArgs(newGasEstimate, newFeedAddr)
      })

      describe('when l1 gas price feed address is the zero address', () => {
        it('reverts', async () => {
          await expect(
            starknetValidator.connect(deployer).setGasConfig(25000, ethers.constants.AddressZero),
          ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
        })
      })
    })

    describe('when access controller address is set', () => {
      describe('when called by an address with access', () => {
        beforeEach(async () => {
          await mockAccessController.mock.hasAccess.returns(true)
        })

        it('correctly sets the gas config', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await starknetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr)
          const gasConfig = await starknetValidator.getGasConfig()
          expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
          expect(gasConfig.gasPriceL1Feed).to.hexEqual(newFeedAddr)
        })

        it('emits an event', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await expect(
            starknetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr),
          )
            .to.emit(starknetValidator, 'GasConfigSet')
            .withArgs(newGasEstimate, newFeedAddr)
        })

        describe('when l1 gas price feed address is the zero address', () => {
          it('reverts', async () => {
            await expect(
              starknetValidator
                .connect(eoaValidator)
                .setGasConfig(25000, ethers.constants.AddressZero),
            ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
          })
        })
      })
    })

    describe('when access controller address is not set', () => {
      beforeEach(async () => {
        await starknetValidator.connect(deployer).setConfigAC(ethers.constants.AddressZero)
      })

      describe('when called by an address without access', () => {
        beforeEach(async () => {
          await mockAccessController.mock.hasAccess.returns(false)
        })

        it('correctly sets the gas config', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await starknetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr)
          const gasConfig = await starknetValidator.getGasConfig()
          expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
          expect(gasConfig.gasPriceL1Feed).to.hexEqual(newFeedAddr)
        })

        it('emits an event', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await expect(
            starknetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr),
          )
            .to.emit(starknetValidator, 'GasConfigSet')
            .withArgs(newGasEstimate, newFeedAddr)
        })

        describe('when l1 gas price feed address is the zero address', () => {
          it('reverts', async () => {
            await expect(
              starknetValidator
                .connect(eoaValidator)
                .setGasConfig(25000, ethers.constants.AddressZero),
            ).to.be.revertedWithCustomError(starknetValidator, 'InvalidGasPriceL1FeedAddress')
          })
        })
      })
    })
  })

  describe('#validate', () => {
    it('reverts if `StarknetValidator.validate` called by account with no access', async () => {
      const c = starknetValidator.connect(eoaValidator)
      await expect(c.validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should not revert if `sequencer_uptime_feed.latest_round_data` called by an Account with no explicit access (Accounts are allowed read access)', async () => {
      const { round } = await l2Contract.call('latest_round_data')
      expect(round.answer).to.equal(0n)
    })

    it('should deploy the messaging contract', async () => {
      const { address, l1_provider } = await starknet.devnet.loadL1MessagingContract(networkUrl)
      expect(address).not.to.be.undefined
      expect(l1_provider).to.equal(networkUrl)
    })

    it('should load the already deployed contract if the address is provided', async () => {
      const { address: loadedFrom } = await starknet.devnet.loadL1MessagingContract(
        networkUrl,
        mockStarknetMessaging.address,
      )

      expect(mockStarknetMessaging.address).to.hexEqual(loadedFrom)
    })

    it('should send a message to the L2 contract', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

      // Simulate L1 transmit + validate
      await starknetValidator.addAccess(eoaValidator.address)
      await starknetValidator.connect(eoaValidator).validate(0, 0, 1, 1)

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1[0].args.to_address).to.hexEqual(l2Contract.address)
      expect(msgFromL1[0].address).to.hexEqual(mockStarknetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(1n)
    })

    it('should always send a **boolean** message to L2 contract', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

      // Simulate L1 transmit + validate
      await starknetValidator.addAccess(eoaValidator.address)
      await starknetValidator.connect(eoaValidator).validate(0, 0, 1, 127) // incorrect value

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1[0].args.to_address).to.hexEqual(l2Contract.address)
      expect(msgFromL1[0].address).to.hexEqual(mockStarknetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(0n) // status unchanged - incorrect value treated as false
    })

    it('should send multiple messages', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

      // Simulate L1 transmit + validate
      await starknetValidator.addAccess(eoaValidator.address)
      const c = starknetValidator.connect(eoaValidator)
      await c.validate(0, 0, 1, 1)
      await c.validate(0, 0, 1, 1)
      await c.validate(0, 0, 1, 127) // incorrect value
      await c.validate(0, 0, 1, 0) // final status

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(4)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.hexEqual(starknetValidator.address)
      expect(msgFromL1[0].args.to_address).to.hexEqual(l2Contract.address)
      expect(msgFromL1[0].address).to.hexEqual(mockStarknetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(0n) // final status 0
    })
  })

  describe('#withdrawFunds', () => {
    beforeEach(async () => {
      await deployer.sendTransaction({ to: starknetValidator.address, value: 10 })
      const balance = await ethers.provider.getBalance(starknetValidator.address)
      expect(balance).to.equal(10n)
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
      await deployer.sendTransaction({ to: starknetValidator.address, value: 10 })
      const balance = await ethers.provider.getBalance(starknetValidator.address)
      expect(balance).to.equal(10n)
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
