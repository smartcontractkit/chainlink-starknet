import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { number } from 'starknet'
import { StarknetContractFactory, StarknetContract, HttpNetworkConfig } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { expectAddressEquality } from './utils'
import { getSelectorFromName } from 'starknet/dist/utils/hash'

describe('StarkNetValidator', () => {
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let starkNetValidator: Contract
  let mockStarkNetMessengerFactory: ContractFactory
  let mockStarkNetMessenger: Contract
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let l2ContractFactory: StarknetContractFactory
  let l2Contract: StarknetContract

  before(async () => {
    const account = await starknet.deployAccount('OpenZeppelin')

    l2ContractFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    l2Contract = await l2ContractFactory.deploy({
      initial_status: 0,
      owner_address: number.toBN(account.starknetContract.address),
    })

    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]

    const starknetValidatorFactory = await ethers.getContractFactory('StarkNetValidator', deployer)
    mockStarkNetMessengerFactory = await ethers.getContractFactory('MockStarkNetMessaging', deployer)

    mockStarkNetMessenger = await mockStarkNetMessengerFactory.deploy()
    await mockStarkNetMessenger.deployed()

    starkNetValidator = await starknetValidatorFactory.deploy(mockStarkNetMessenger.address, l2Contract.address)

    await account.invoke(l2Contract, 'set_l1_sender', { address: starkNetValidator.address })
  })

  describe('starknetValidator', () => {
    it('should get the selector from name successfully', async () => {
      const setSelector = getSelectorFromName('update_status')
      expect(BigInt(setSelector)).to.equal(
        1585322027166395525705364165097050997465692350398750944680096081848180365267n,
      )
    })

    it('reverts if called by an account with no access', async () => {
      await expect(starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should deploy the messaging contract', async () => {
      starkNetValidator.addAccess(eoaValidator.address)

      const { address: deployedTo, l1_provider: L1Provider } = await starknet.devnet.loadL1MessagingContract(networkUrl)

      expect(deployedTo).not.to.be.undefined
      expect(L1Provider).to.equal(networkUrl)

      const { address: loadedFrom } = await starknet.devnet.loadL1MessagingContract(
        networkUrl,
        mockStarkNetMessenger.address,
      )
      expect(mockStarkNetMessenger.address).to.equal(loadedFrom)

      await starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 1)
      const flushL1Response = await starknet.devnet.flush()
      const flushL1Messages = flushL1Response.consumed_messages.from_l1
      expect(flushL1Messages).to.have.a.lengthOf(1)
      expect(flushL1Response.consumed_messages.from_l2).to.be.empty
      expectAddressEquality(flushL1Messages[0].args.from_address, starkNetValidator.address)
      expectAddressEquality(flushL1Messages[0].args.to_address, l2Contract.address)
      expectAddressEquality(flushL1Messages[0].address, mockStarkNetMessenger.address)
    })
  })
})
