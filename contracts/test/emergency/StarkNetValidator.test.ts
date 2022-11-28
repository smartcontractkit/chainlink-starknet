import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { number } from 'starknet'
import {
  Account,
  StarknetContractFactory,
  StarknetContract,
  HttpNetworkConfig,
} from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { getSelectorFromName } from 'starknet/dist/utils/hash'
import { abi as aggregatorAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol/AggregatorV3Interface.json'
import { abi as accessControllerAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol/AccessControllerInterface.json'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'
import {
  account,
  addCompilationToNetwork,
  loadConfig,
  NetworkManager,
  FunderOptions,
  Funder,
} from '@chainlink/starknet'

describe('StarkNetValidator', () => {
  const config = loadConfig()
  const optsConf = { config, required: ['devnet', 'hardhat'] }
  const manager = new NetworkManager(optsConf)
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let opts: FunderOptions
  let funder: Funder

  let defaultAccount: Account
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress
  let alice: SignerWithAddress

  let starkNetValidator: Contract
  let mockStarkNetMessagingFactory: ContractFactory
  let mockStarkNetMessaging: Contract
  let mockGasPriceFeed: MockContract
  let mockAccessController: MockContract
  let mockAggregator: MockContract

  let l2ContractFactory: StarknetContractFactory
  let l2Contract: StarknetContract

  before(async () => {
    await manager.start()
    opts = account.makeFunderOptsFromEnv()
    funder = new account.Funder(opts)
    await addCompilationToNetwork(
      'src/chainlink/solidity/emergency/StarkNetValidator.sol:StarkNetValidator',
    )

    // Deploy L2 account
    defaultAccount = await starknet.deployAccount('OpenZeppelin')

    // Fund L2 account
    await funder.fund([{ account: defaultAccount.address, amount: 5000 }])

    // Fetch predefined L1 EOA accounts
    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]
    alice = accounts[2]

    // Deploy L2 feed contract
    l2ContractFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    l2Contract = await l2ContractFactory.deploy({
      initial_status: 0,
      owner_address: number.toBN(defaultAccount.starknetContract.address),
    })

    // Deploy the MockStarkNetMessaging contract used to simulate L1 - L2 comms
    mockStarkNetMessagingFactory = await ethers.getContractFactory(
      'MockStarkNetMessaging',
      deployer,
    )
    mockStarkNetMessaging = await mockStarkNetMessagingFactory.deploy()
    await mockStarkNetMessaging.deployed()

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
    // Deploy the L1 StarkNetValidator
    const starknetValidatorFactory = await ethers.getContractFactory('StarkNetValidator', deployer)
    starkNetValidator = await starknetValidatorFactory.deploy(
      mockStarkNetMessaging.address,
      mockAccessController.address,
      mockGasPriceFeed.address,
      mockAggregator.address,
      l2Contract.address,
      0,
    )

    // Point the L2 feed contract to receive from the L1 StarkNetValidator contract
    await defaultAccount.invoke(l2Contract, 'set_l1_sender', { address: starkNetValidator.address })
  })

  describe('#constructor', () => {
    it('reverts when the StarknetMessaging address is zero', async () => {
      const factory = await ethers.getContractFactory('StarkNetValidator', deployer)
      await expect(
        factory.deploy(
          ethers.constants.AddressZero,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          l2Contract.address,
          0,
        ),
      ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidStarkNetMessagingAddress')
    })

    it('reverts when the L2 feed is zero', async () => {
      const factory = await ethers.getContractFactory('StarkNetValidator', deployer)
      await expect(
        factory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          mockAggregator.address,
          0,
          0,
        ),
      ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidL2FeedAddress')
    })

    it('reverts when the Aggregator address is zero', async () => {
      const factory = await ethers.getContractFactory('StarkNetValidator', deployer)
      await expect(
        factory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          mockGasPriceFeed.address,
          ethers.constants.AddressZero,
          l2Contract.address,
          0,
        ),
      ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidSourceAggregatorAddress')
    })

    it('reverts when the L1 Gas Price feed address is zero', async () => {
      const factory = await ethers.getContractFactory('StarkNetValidator', deployer)
      await expect(
        factory.deploy(
          mockStarkNetMessaging.address,
          mockAccessController.address,
          ethers.constants.AddressZero,
          mockAggregator.address,
          l2Contract.address,
          0,
        ),
      ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidGasPriceL1FeedAddress')
    })

    it('is initialized with the correct gas config', async () => {
      const gasConfig = await starkNetValidator.getGasConfig()
      expect(gasConfig.gasEstimate).to.equal(0) // Initialized with 0 in before function
      expect(gasConfig.gasPriceL1Feed).to.equal(mockGasPriceFeed.address)
    })

    it('is initialized with the correct access controller address', async () => {
      const acAddr = await starkNetValidator.getConfigAC()
      expect(acAddr).to.equal(mockAccessController.address)
    })

    it('is initialized with the correct source aggregator address', async () => {
      const aggregatorAddr = await starkNetValidator.getSourceAggregator()
      expect(aggregatorAddr).to.equal(mockAggregator.address)
    })

    it('should get the selector from the name successfully', async () => {
      const actual = getSelectorFromName('update_status')
      const expected = 1585322027166395525705364165097050997465692350398750944680096081848180365267n
      expect(BigInt(actual)).to.equal(expected)

      const computedActual = await starkNetValidator.SELECTOR_STARK_UPDATE_STATUS()
      expect(BigInt(computedActual)).to.equal(expected)
    })
  })

  describe('#retry', () => {
    describe('when called by account with no access', () => {
      it('reverts', async () => {
        await expect(starkNetValidator.connect(alice).retry()).to.be.revertedWith('No access')
      })
    })
  })

  describe('#setConfigAC', () => {
    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(alice).setConfigAC(ethers.constants.AddressZero),
        ).to.be.revertedWith('Only callable by owner')
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        const newACAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starkNetValidator.setConfigAC(newACAddr))
          .to.emit(starkNetValidator, 'ConfigACSet')
          .withArgs(mockAccessController.address, newACAddr)
      })

      it('sets the access controller address', async () => {
        const newACAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await starkNetValidator.connect(deployer).setConfigAC(newACAddr)
        expect(await starkNetValidator.getConfigAC()).to.equal(newACAddr)
      })
    })
  })

  describe('#setSourceAggregator', () => {
    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(alice).setSourceAggregator(ethers.constants.AddressZero),
        ).to.be.revertedWith('Only callable by owner')
      })
    })

    describe('when source address is the zero address', () => {
      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(deployer).setSourceAggregator(ethers.constants.AddressZero),
        ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidSourceAggregatorAddress')
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        const newSourceAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starkNetValidator.setSourceAggregator(newSourceAddr))
          .to.emit(starkNetValidator, 'SourceAggregatorSet')
          .withArgs(mockAggregator.address, newSourceAddr)
      })

      it('sets the source aggregator address', async () => {
        await starkNetValidator.connect(deployer).setSourceAggregator(mockAggregator.address)
        expect(await starkNetValidator.getSourceAggregator()).to.equal(mockAggregator.address)
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
          starkNetValidator.connect(alice).setGasConfig(0, mockGasPriceFeed.address),
        ).to.be.revertedWithCustomError(starkNetValidator, 'AccessForbidden')
      })
    })

    describe('when called by owner', () => {
      it('correctly sets the gas config', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await starkNetValidator.connect(deployer).setGasConfig(newGasEstimate, newFeedAddr)
        const gasConfig = await starkNetValidator.getGasConfig()
        expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
        expect(gasConfig.gasPriceL1Feed).to.equal(newFeedAddr)
      })

      it('emits an event', async () => {
        const newGasEstimate = 25000
        const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
        await expect(starkNetValidator.connect(deployer).setGasConfig(newGasEstimate, newFeedAddr))
          .to.emit(starkNetValidator, 'GasConfigSet')
          .withArgs(newGasEstimate, newFeedAddr)
      })

      describe('when l1 gas price feed address is the zero address', () => {
        it('reverts', async () => {
          await expect(
            starkNetValidator.connect(deployer).setGasConfig(25000, ethers.constants.AddressZero),
          ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidGasPriceL1FeedAddress')
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
          await starkNetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr)
          const gasConfig = await starkNetValidator.getGasConfig()
          expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
          expect(gasConfig.gasPriceL1Feed).to.equal(newFeedAddr)
        })

        it('emits an event', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await expect(
            starkNetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr),
          )
            .to.emit(starkNetValidator, 'GasConfigSet')
            .withArgs(newGasEstimate, newFeedAddr)
        })

        describe('when l1 gas price feed address is the zero address', () => {
          it('reverts', async () => {
            await expect(
              starkNetValidator
                .connect(eoaValidator)
                .setGasConfig(25000, ethers.constants.AddressZero),
            ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidGasPriceL1FeedAddress')
          })
        })
      })
    })

    describe('when access controller address is not set', () => {
      beforeEach(async () => {
        await starkNetValidator.connect(deployer).setConfigAC(ethers.constants.AddressZero)
      })

      describe('when called by an address without access', () => {
        beforeEach(async () => {
          await mockAccessController.mock.hasAccess.returns(false)
        })

        it('correctly sets the gas config', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await starkNetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr)
          const gasConfig = await starkNetValidator.getGasConfig()
          expect(gasConfig.gasEstimate).to.equal(newGasEstimate)
          expect(gasConfig.gasPriceL1Feed).to.equal(newFeedAddr)
        })

        it('emits an event', async () => {
          const newGasEstimate = 25000
          const newFeedAddr = '0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4'
          await expect(
            starkNetValidator.connect(eoaValidator).setGasConfig(newGasEstimate, newFeedAddr),
          )
            .to.emit(starkNetValidator, 'GasConfigSet')
            .withArgs(newGasEstimate, newFeedAddr)
        })

        describe('when l1 gas price feed address is the zero address', () => {
          it('reverts', async () => {
            await expect(
              starkNetValidator
                .connect(eoaValidator)
                .setGasConfig(25000, ethers.constants.AddressZero),
            ).to.be.revertedWithCustomError(starkNetValidator, 'InvalidGasPriceL1FeedAddress')
          })
        })
      })
    })
  })

  describe('#validate', () => {
    it('reverts if `StarkNetValidator.validate` called by account with no access', async () => {
      const c = starkNetValidator.connect(eoaValidator)
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
        mockStarkNetMessaging.address,
      )

      expect(mockStarkNetMessaging.address).to.equal(loadedFrom)
    })

    it('should send a message to the L2 contract', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessaging.address)

      // Simulate L1 transmit + validate
      await starkNetValidator.addAccess(eoaValidator.address)
      await starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 1)

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.equal(starkNetValidator.address)
      expect(msgFromL1[0].args.to_address).to.equal(l2Contract.address)
      expect(msgFromL1[0].address).to.equal(mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(1n)
    })

    it('should always send a **boolean** message to L2 contract', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessaging.address)

      // Simulate L1 transmit + validate
      await starkNetValidator.addAccess(eoaValidator.address)
      await starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 127) // incorrect value

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.equal(starkNetValidator.address)
      expect(msgFromL1[0].args.to_address).to.equal(l2Contract.address)
      expect(msgFromL1[0].address).to.equal(mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(0n) // status unchanged - incorrect value treated as false
    })

    it('should send multiple messages', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessaging.address)

      // Simulate L1 transmit + validate
      await starkNetValidator.addAccess(eoaValidator.address)
      const c = starkNetValidator.connect(eoaValidator)
      await c.validate(0, 0, 1, 1)
      await c.validate(0, 0, 1, 1)
      await c.validate(0, 0, 1, 127) // incorrect value
      await c.validate(0, 0, 1, 0) // final status

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(4)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expect(msgFromL1[0].args.from_address).to.equal(starkNetValidator.address)
      expect(msgFromL1[0].args.to_address).to.equal(l2Contract.address)
      expect(msgFromL1[0].address).to.equal(mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(0n) // final status 0
    })
  })

  describe('#withdrawFunds', () => {
    beforeEach(async () => {
      await deployer.sendTransaction({ to: starkNetValidator.address, value: 10 })
      const balance = await ethers.provider.getBalance(starkNetValidator.address)
      expect(balance).to.equal(10n)
    })

    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(starkNetValidator.connect(alice).withdrawFunds()).to.be.revertedWith(
          'Only callable by owner',
        )
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        await expect(starkNetValidator.connect(deployer).withdrawFunds())
          .to.emit(starkNetValidator, 'FundsWithdrawn')
          .withArgs(deployer.address, 10)
      })

      it('withdraws all funds to deployer', async () => {
        await starkNetValidator.connect(deployer).withdrawFunds()
        const balance = await ethers.provider.getBalance(starkNetValidator.address)
        expect(balance).to.equal(0n)
      })
    })
  })

  describe('#withdrawFundsTo', () => {
    beforeEach(async () => {
      await deployer.sendTransaction({ to: starkNetValidator.address, value: 10 })
      const balance = await ethers.provider.getBalance(starkNetValidator.address)
      expect(balance).to.equal(10n)
    })

    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(alice).withdrawFundsTo(alice.address),
        ).to.be.revertedWith('Only callable by owner')
      })
    })

    describe('when called by owner', () => {
      it('emits an event', async () => {
        await expect(starkNetValidator.connect(deployer).withdrawFundsTo(eoaValidator.address))
          .to.emit(starkNetValidator, 'FundsWithdrawn')
          .withArgs(eoaValidator.address, 10)
      })

      it('withdraws all funds to deployer', async () => {
        await starkNetValidator.connect(deployer).withdrawFunds()
        const balance = await ethers.provider.getBalance(starkNetValidator.address)
        expect(balance).to.equal(0n)
      })
    })
  })

  after(async function () {
    manager.stop()
  })
})
