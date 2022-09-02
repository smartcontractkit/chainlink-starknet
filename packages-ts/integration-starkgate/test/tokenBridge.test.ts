import { TIMEOUT } from './constants'
import { ethers, starknet, network } from 'hardhat'
import { Contract, ContractFactory } from 'ethers'
import { uint256, number } from 'starknet'
import { StarknetContract, HttpNetworkConfig, Account } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import { expectAddressEquality } from '../../../contracts/test/utils'
import {
  loadContract_Solidity,
  loadContract_InternalStarkgate,
  loadContract_Openzepplin,
} from '../utils'

const NAME = 'ChainLink Token'
const SYMBOL = 'LINK'

describe('Test starkgate bridge with link token', function () {
  this.timeout(TIMEOUT)
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let owner: Account
  let tokenBridgeContract: StarknetContract
  let linkTokenContract: StarknetContract
  let deployer: SignerWithAddress

  let proxyFactory: ContractFactory

  let starkNetERC20Bridge: Contract
  let mockStarknetMessaging: Contract
  let proxy: Contract
  let testERC20: Contract
  let newStarkNetERC20Bridge: Contract

  let starkNetERC20BridgeContract: any

  before(async () => {
    owner = await starknet.deployAccount('OpenZeppelin')

    let tokenBridgeFactory = await starknet.getContractFactory(
      '../../node_modules/@chainlink-dev/starkgate-contracts/artifacts/token_bridge.cairo',
    )
    tokenBridgeContract = await tokenBridgeFactory.deploy({
      governor_address: owner.starknetContract.address,
    })

    let linkTokenFactory = await starknet.getContractFactory('link_token')
    linkTokenContract = await linkTokenFactory.deploy({ owner: tokenBridgeContract.address })

    const accounts = await ethers.getSigners()
    deployer = accounts[0]

    starkNetERC20BridgeContract = await loadContract_InternalStarkgate('StarknetERC20Bridge')
    const starkNetERC20BridgeFactory = new ethers.ContractFactory(
      starkNetERC20BridgeContract.abi,
      starkNetERC20BridgeContract.bytecode,
      deployer,
    )
    starkNetERC20Bridge = await starkNetERC20BridgeFactory.deploy()
    await starkNetERC20Bridge.deployed()

    const mockStarknetMessagingContract = await loadContract_Solidity('MockStarkNetMessaging')
    const mockStarknetMessagingFactory = new ethers.ContractFactory(
      mockStarknetMessagingContract.abi,
      mockStarknetMessagingContract.bytecode,
      deployer,
    )
    mockStarknetMessaging = await mockStarknetMessagingFactory.deploy()
    await mockStarknetMessaging.deployed()

    await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

    const testERC20contract = await loadContract_Openzepplin('ERC20PresetFixedSupply')
    const testERC20Factory = new ethers.ContractFactory(
      testERC20contract.abi,
      testERC20contract.bytecode,
      deployer,
    )
    testERC20 = await testERC20Factory.deploy(NAME, SYMBOL, 10, deployer.address)
    await testERC20.deployed()

    const proxyContract = await loadContract_Openzepplin('ERC1967Proxy')
    proxyFactory = new ethers.ContractFactory(proxyContract.abi, proxyContract.bytecode, deployer)

    const inter = new ethers.utils.Interface(starkNetERC20BridgeContract.abi)
    const data = ethers.utils.hexConcat([
      ethers.utils.hexZeroPad(ethers.constants.AddressZero, 32),
      ethers.utils.hexZeroPad(testERC20.address, 32),
      ethers.utils.hexZeroPad(mockStarknetMessaging.address, 32),
    ])
    let encode_data = inter.encodeFunctionData('initialize(bytes data)', [data])
    proxy = await proxyFactory.deploy(starkNetERC20Bridge.address, encode_data)
    await proxy.deployed()
  })

  describe('Test bridge from L1 to L2', function () {
    it('Test Set and Get function for L2 token address', async () => {
      await owner.invoke(tokenBridgeContract, 'set_l2_token', {
        l2_token_address: linkTokenContract.address,
      })
      const { res: l2_address } = await tokenBridgeContract.call('get_l2_token', {})
      expectAddressEquality(l2_address.toString(), linkTokenContract.address)
    })

    it('Should wrap contract and set L2 TokenBridge successfully', async () => {
      newStarkNetERC20Bridge = await ethers.getContractAt(
        starkNetERC20BridgeContract.abi,
        proxy.address,
      )

      const tx = await newStarkNetERC20Bridge.setL2TokenBridge(BigInt(tokenBridgeContract.address))
      await expect(tx)
        .to.emit(newStarkNetERC20Bridge, 'LogSetL2TokenBridge')
        .withArgs(BigInt(tokenBridgeContract.address))
    })

    it('Test Set and Get function for L1 bridge address', async () => {
      await owner.invoke(tokenBridgeContract, 'set_l1_bridge', {
        l1_bridge_address: newStarkNetERC20Bridge.address,
      })
      const { res: l1_address } = await tokenBridgeContract.call('get_l1_bridge', {})
      expectAddressEquality(l1_address.toString(), newStarkNetERC20Bridge.address)
    })

    it('Should set Max total balance', async () => {
      await newStarkNetERC20Bridge.setMaxTotalBalance(100000)
      const totalbalance = await newStarkNetERC20Bridge.maxTotalBalance()
      expect(totalbalance).to.equal(100000)
    })

    it('Should set Max deposit', async () => {
      await newStarkNetERC20Bridge.setMaxDeposit(100)
      const deposit = await newStarkNetERC20Bridge.maxDeposit()
      expect(deposit).to.equal(100)
    })

    it('Should deposit to the L2 contract, L1 balance should be decreased by 2', async () => {
      await testERC20.approve(newStarkNetERC20Bridge.address, 2)
      await newStarkNetERC20Bridge.deposit(2, owner.starknetContract.address)

      const balance = await testERC20.balanceOf(deployer.address)
      expect(balance).to.equal(8)
    })

    it('Should flush the L1 messages so that they can be consumed by the L2.', async () => {
      const flushL1Response = await starknet.devnet.flush()
      const flushL1Messages = flushL1Response.consumed_messages.from_l1
      expect(flushL1Messages).to.have.a.lengthOf(1)
      expect(flushL1Response.consumed_messages.from_l2).to.be.empty

      expectAddressEquality(flushL1Messages[0].args.from_address, newStarkNetERC20Bridge.address)
      expectAddressEquality(flushL1Messages[0].args.to_address, tokenBridgeContract.address)
      expectAddressEquality(flushL1Messages[0].address, mockStarknetMessaging.address)

      let { balance: balance } = await linkTokenContract.call('balanceOf', {
        account: owner.starknetContract.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(number.toBN(2))
    })
  })

  describe('Test bridge from L2 to L1', function () {
    it('Should initiate withdraw and send message to L1', async () => {
      await owner.invoke(tokenBridgeContract, 'initiate_withdraw', {
        l1_recipient: BigInt(deployer.address),
        amount: uint256.bnToUint256(2),
      })
      let { balance: balance } = await linkTokenContract.call('balanceOf', {
        account: owner.starknetContract.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(number.toBN(0))

      const flushL2Response = await starknet.devnet.flush()
      expect(flushL2Response.consumed_messages.from_l1).to.be.empty
      const flushL2Messages = flushL2Response.consumed_messages.from_l2

      expect(flushL2Messages).to.have.a.lengthOf(1)
      expectAddressEquality(flushL2Messages[0].from_address, tokenBridgeContract.address)
      expectAddressEquality(flushL2Messages[0].to_address, newStarkNetERC20Bridge.address)
    })

    it('Should withdraw 2 which will consume the L2 message successfully', async () => {
      await newStarkNetERC20Bridge['withdraw(uint256)'](2)

      const balance = await testERC20.balanceOf(deployer.address)
      expect(balance).to.equal(10)
    })
  })
})
