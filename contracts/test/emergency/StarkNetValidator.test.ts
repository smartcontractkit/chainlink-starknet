import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { number } from 'starknet'
import { StarknetContractFactory, StarknetContract, HttpNetworkConfig } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { expectAddressEquality } from '../utils'
import { getSelectorFromName } from 'starknet/dist/utils/hash'

describe('StarkNetValidator', () => {
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url

  let starkNetValidator: Contract
  let mockStarkNetMessagingFactory: ContractFactory
  let mockStarkNetMessaging: Contract
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
    mockStarkNetMessagingFactory = await ethers.getContractFactory(
      'MockStarkNetMessaging',
      deployer,
    )

    mockStarkNetMessaging = await mockStarkNetMessagingFactory.deploy()
    await mockStarkNetMessaging.deployed()

    starkNetValidator = await starknetValidatorFactory.deploy(
      mockStarkNetMessaging.address,
      l2Contract.address,
    )

    await account.invoke(l2Contract, 'set_l1_sender', { address: starkNetValidator.address })
  })

  describe('StarkNetValidator', () => {
    it('should get the selector from the name successfully', async () => {
      const setSelector = getSelectorFromName('update_status')
      expect(BigInt(setSelector)).to.equal(
        1585322027166395525705364165097050997465692350398750944680096081848180365267n,
      )
    })

    it('reverts if called by account with no access', async () => {
      await expect(starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 1)).to.be.revertedWith(
        'No access',
      )
    })

    it('should deploy the messaging contract', async () => {
      const {
        address: deployedTo,
        l1_provider: L1Provider,
      } = await starknet.devnet.loadL1MessagingContract(networkUrl)

      expect(deployedTo).not.to.be.undefined
      expect(L1Provider).to.equal(networkUrl)
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

      expectAddressEquality(msgFromL1[0].args.from_address, starkNetValidator.address)
      expectAddressEquality(msgFromL1[0].args.to_address, l2Contract.address)
      expectAddressEquality(msgFromL1[0].address, mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(1n)
    })

    it('should always send a **boolean** message to L2 contract', async () => {
      // Load the mock messaging contract
      await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarkNetMessaging.address);

      // Simulate L1 transmit + validate
      await starkNetValidator.addAccess(eoaValidator.address)
      await starkNetValidator.connect(eoaValidator).validate(0, 0, 1, 127) // incorrect value

      // Simulate the L1 - L2 comms
      const resp = await starknet.devnet.flush()
      const msgFromL1 = resp.consumed_messages.from_l1
      expect(msgFromL1).to.have.a.lengthOf(1)
      expect(resp.consumed_messages.from_l2).to.be.empty

      expectAddressEquality(msgFromL1[0].args.from_address, starkNetValidator.address)
      expectAddressEquality(msgFromL1[0].args.to_address, l2Contract.address)
      expectAddressEquality(msgFromL1[0].address, mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call("latest_round_data");
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

      expectAddressEquality(msgFromL1[0].args.from_address, starkNetValidator.address)
      expectAddressEquality(msgFromL1[0].args.to_address, l2Contract.address)
      expectAddressEquality(msgFromL1[0].address, mockStarkNetMessaging.address)

      // Assert L2 effects
      const res = await l2Contract.call('latest_round_data')
      expect(res.round.answer).to.equal(0n) // final status 0
    })
  })
})
