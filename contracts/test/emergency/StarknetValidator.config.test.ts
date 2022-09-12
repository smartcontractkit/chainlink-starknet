import { ethers } from 'hardhat'
import { Contract } from 'ethers'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { abi as aggregatorAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol/AggregatorV3Interface.json'
import { abi as accessControllerAbi } from '../../artifacts/@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol/AccessControllerInterface.json'
import { deployMockContract, MockContract } from '@ethereum-waffle/mock-contract'

describe('StarkNetValidator (config)', () => {
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let starkNetValidator: Contract
  let mockGasPriceFeed: MockContract
  let mockAccessController: MockContract
  let mockAggregator: MockContract

  beforeEach(async () => {
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

    // Deploy the L1 StarkNetValidator
    const starknetValidatorFactory = await ethers.getContractFactory('StarkNetValidator', deployer)
    starkNetValidator = await starknetValidatorFactory
      .connect(deployer)
      .deploy(
        '0xde29d060D45901Fb19ED6C6e959EB22d8626708e',
        mockAccessController.address,
        mockGasPriceFeed.address,
        mockAggregator.address,
        '0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f',
        0,
      )
  })

  describe('#retry', () => {
    describe('when called by account with no access', () => {
      it('reverts', async () => {
        await expect(starkNetValidator.connect(eoaValidator).retry()).to.be.revertedWith(
          'No access',
        )
      })
    })
  })

  describe('#setConfigAC', () => {
    describe('when called by non owner', () => {
      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(eoaValidator).setConfigAC(ethers.constants.AddressZero),
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

  describe('#setGasConfig', () => {
    describe('when called by non owner without access', () => {
      beforeEach(async () => {
        await mockAccessController.mock.hasAccess.returns(false)
      })

      it('reverts', async () => {
        await expect(
          starkNetValidator.connect(eoaValidator).setGasConfig(0, mockGasPriceFeed.address),
        ).to.be.revertedWith('AccessForbidden()')
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
          ).to.be.revertedWith('InvalidGasPriceL1FeedAddress()')
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
            ).to.be.revertedWith('InvalidGasPriceL1FeedAddress()')
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
            ).to.be.revertedWith('InvalidGasPriceL1FeedAddress()')
          })
        })
      })
    })
  })
})
