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
import { TIMEOUT } from '../constants'

describe('StarkNetValidator', function () {
  this.timeout(TIMEOUT)

  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url

  let account: Account
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let starkNetValidator: Contract
  let mockStarkNetMessagingFactory: ContractFactory
  let mockStarkNetMessaging: Contract

  let l2ContractFactory: StarknetContractFactory
  let l2Contract: StarknetContract

  before(async () => {
    // Deploy L2 account
    account = await starknet.deployAccount('OpenZeppelin')
    // Fetch predefined L1 EOA accounts
    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]

    // Deploy L2 feed contract
    l2ContractFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    l2Contract = await l2ContractFactory.deploy({
      initial_status: 0,
      owner_address: number.toBN(account.starknetContract.address),
    })

    // Deploy the MockStarkNetMessaging contract used to simulate L1 - L2 comms
    mockStarkNetMessagingFactory = await ethers.getContractFactory(
      'MockStarkNetMessaging',
      deployer,
    )
    mockStarkNetMessaging = await mockStarkNetMessagingFactory.deploy()
    await mockStarkNetMessaging.deployed()

    // Deploy the L1 StarkNetValidator
    const starknetValidatorFactory = await ethers.getContractFactory('StarkNetValidator', deployer)
    starkNetValidator = await starknetValidatorFactory.deploy(
      mockStarkNetMessaging.address,
      l2Contract.address,
    )

    // Point the L2 feed contract to receive from the L1 StarkNetValidator contract
    await account.invoke(l2Contract, 'set_l1_sender', { address: starkNetValidator.address })
  })

  describe('StarkNetValidator', () => {
    it('should get the selector from the name successfully', async () => {
      const actual = getSelectorFromName('update_status')
      const expected = 1585322027166395525705364165097050997465692350398750944680096081848180365267n
      expect(BigInt(actual)).to.equal(expected)
    })

    it('reverts if `StarkNetValidator.validate` called by account with no access', async () => {
      const c = starkNetValidator.connect(eoaValidator)
      await expect(c.validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should not revert if `sequencer_uptime_feed.latest_round_data` called by an Account with no explicit access (Accounts are allowed read access)', async () => {
      const { round } = await account.call(l2Contract, 'latest_round_data')
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
      const res = await account.call(l2Contract, 'latest_round_data')
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
      const res = await account.call(l2Contract, 'latest_round_data')
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
      const res = await account.call(l2Contract, 'latest_round_data')
      expect(res.round.answer).to.equal(0n) // final status 0
    })
  })
})
