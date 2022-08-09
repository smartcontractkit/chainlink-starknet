import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { number } from 'starknet'
import { StarknetContractFactory, StarknetContract, HttpNetworkConfig } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { expectAddressEquality } from './utils'
import { getSelectorFromName } from 'starknet/dist/utils/hash'

describe('StarknetValidator', () => {
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let Validator: Contract
  let MockStarknetMessaging: ContractFactory
  let mockStarknetMessaging: Contract
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let L2contractFactory: StarknetContractFactory
  let l2contract: StarknetContract

  before(async () => {
    const account = await starknet.deployAccount('OpenZeppelin')

    L2contractFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    l2contract = await L2contractFactory.deploy({
      initial_status: 0,
      owner_address: number.toBN(account.starknetContract.address),
    })

    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]

    const ValidatorFactory = await ethers.getContractFactory('Validator', deployer)
    MockStarknetMessaging = await ethers.getContractFactory('MockStarknetMessaging', deployer)

    mockStarknetMessaging = await MockStarknetMessaging.deploy()
    await mockStarknetMessaging.deployed()

    Validator = await ValidatorFactory.deploy(mockStarknetMessaging.address, l2contract.address)

    await account.invoke(l2contract, 'set_l1_sender', { address: Validator.address })
  })

  describe('#validate', () => {
    it('should get the selector from name successfully', async () => {
      const setSelector = getSelectorFromName('update_status')
      expect(BigInt(setSelector).toString(10)).to.equal(1585322027166395525705364165097050997465692350398750944680096081848180365267n)
    })

    it('reverts if called by account with no access', async () => {
      await expect(Validator.connect(eoaValidator).validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should connect successfully', async () => {
      Validator.addAccess(eoaValidator.address)
      Validator.connect(eoaValidator).validate(0, 0, 1, 1)
    })

    it('should deploy the messaging contract', async () => {
      Validator.addAccess(eoaValidator.address)

      const { address: deployedTo, l1_provider: L1Provider } = await starknet.devnet.loadL1MessagingContract(networkUrl)

      expect(deployedTo).not.to.be.undefined
      expect(L1Provider).to.equal(networkUrl)

      const { address: loadedFrom } = await starknet.devnet.loadL1MessagingContract(
        networkUrl,
        mockStarknetMessaging.address,
      )

      expect(mockStarknetMessaging.address).to.equal(loadedFrom)

      await Validator.connect(eoaValidator).validate(0, 0, 1, 1)

      const flushL1Response = await starknet.devnet.flush()
      const flushL1Messages = flushL1Response.consumed_messages.from_l1
      expect(flushL1Messages).to.have.a.lengthOf(1)
      expect(flushL1Response.consumed_messages.from_l2).to.be.empty

      expectAddressEquality(flushL1Messages[0].args.from_address, Validator.address)
      expectAddressEquality(flushL1Messages[0].args.to_address, l2contract.address)
      expectAddressEquality(flushL1Messages[0].address, mockStarknetMessaging.address)
    })
  })
})
