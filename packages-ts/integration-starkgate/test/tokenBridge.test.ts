import { TIMEOUT } from './constants'
import { ethers, starknet, network } from 'hardhat'
import { Contract } from 'ethers'
import { uint256, number } from 'starknet'
import { StarknetContract, HttpNetworkConfig, Account } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
import {
  loadContract_Solidity,
  loadContract_InternalStarkgate,
  loadContract_OpenZepplin,
} from '../src/utils'
import {
  account,
  loadConfig,
  hexPadStart,
  NetworkManager,
  FunderOptions,
  Funder,
} from '@chainlink/starknet'

const NAME = 'ChainLink Token'
const SYMBOL = 'LINK'

describe('Test StarkGate token bridge + link_token.cairo', function () {
  this.timeout(TIMEOUT)
  const config = loadConfig()
  const optsConf = { config, required: ['starknet', 'ethereum'] }
  const manager = new NetworkManager(optsConf)

  let opts: FunderOptions
  let funder: Funder

  // L2 StarkNet
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let owner: Account
  let tokenBridge: StarknetContract
  let tokenL2: StarknetContract

  // L1 Ethereum
  let deployer: SignerWithAddress
  let mockStarknetMessaging: Contract
  let tokenL1: Contract
  let starknetERC20Bridge: Contract

  before(async () => {
    await manager.start()

    opts = account.makeFunderOptsFromEnv()
    funder = new account.Funder(opts)

    owner = await starknet.deployAccount('OpenZeppelin')

    await funder.fund([{ account: owner.address, amount: 5000 }])

    let tokenBridgeFactory = await starknet.getContractFactory(
      '../../node_modules/@chainlink-dev/starkgate-contracts/artifacts/token_bridge.cairo',
    )
    tokenBridge = await tokenBridgeFactory.deploy({
      governor_address: owner.starknetContract.address,
    })

    let linkTokenFactory = await starknet.getContractFactory('link_token')
    tokenL2 = await linkTokenFactory.deploy({ owner: tokenBridge.address })
    ;[deployer] = await ethers.getSigners()
    // Different .json artifact - incompatible with 'ethers.getContractFactoryFromArtifact'
    const starknetERC20BridgeArtifact = await loadContract_InternalStarkgate('StarknetERC20Bridge')
    const starkNetERC20BridgeFactory = new ethers.ContractFactory(
      starknetERC20BridgeArtifact.abi,
      starknetERC20BridgeArtifact.bytecode,
      deployer,
    )
    const starkNetERC20BridgeCode = await starkNetERC20BridgeFactory.deploy()
    await starkNetERC20BridgeCode.deployed()

    const mockStarknetMessagingArtifact = await loadContract_Solidity('MockStarkNetMessaging')
    const mockStarknetMessagingFactory = await ethers.getContractFactoryFromArtifact(
      mockStarknetMessagingArtifact,
      deployer,
    )
    mockStarknetMessaging = await mockStarknetMessagingFactory.deploy()
    await mockStarknetMessaging.deployed()

    await starknet.devnet.loadL1MessagingContract(networkUrl, mockStarknetMessaging.address)

    const tokenERC20Artifact = await loadContract_OpenZepplin('ERC20PresetFixedSupply')
    const tokenERC20Factory = await ethers.getContractFactoryFromArtifact(
      tokenERC20Artifact,
      deployer,
    )
    tokenL1 = await tokenERC20Factory.deploy(NAME, SYMBOL, 10, deployer.address)
    await tokenL1.deployed()

    const proxyArtifact = await loadContract_OpenZepplin('ERC1967Proxy')
    const proxyFactory = await ethers.getContractFactoryFromArtifact(proxyArtifact, deployer)

    const inter = new ethers.utils.Interface(starknetERC20BridgeArtifact.abi)
    const data = ethers.utils.hexConcat([
      ethers.utils.hexZeroPad(ethers.constants.AddressZero, 32),
      ethers.utils.hexZeroPad(tokenL1.address, 32),
      ethers.utils.hexZeroPad(mockStarknetMessaging.address, 32),
    ])
    let encode_data = inter.encodeFunctionData('initialize(bytes data)', [data])
    const proxy = await proxyFactory.deploy(starkNetERC20BridgeCode.address, encode_data)
    await proxy.deployed()

    starknetERC20Bridge = await ethers.getContractAt(starknetERC20BridgeArtifact.abi, proxy.address)
  })

  describe('deposit (L1 -> L2)', function () {
    it('should configure L2 token bridge with L2 token successfully', async () => {
      await owner.invoke(tokenBridge, 'set_l2_token', {
        l2_token_address: tokenL2.address,
      })
      const { res: l2Token } = await tokenBridge.call('get_l2_token', {})
      expect(hexPadStart(l2Token, 32 * 2)).to.equal(tokenL2.address)
    })

    it(`should configure L2 token bridge with L1 token bridge successfully`, async () => {
      await owner.invoke(tokenBridge, 'set_l1_bridge', {
        l1_bridge_address: starknetERC20Bridge.address,
      })
      const { res: l1Bridge } = await tokenBridge.call('get_l1_bridge', {})
      expect(hexPadStart(l1Bridge, 20 * 2)).to.equal(starknetERC20Bridge.address.toLowerCase())
    })

    it('should configure L1 token bridge with L2 token bridge successfully', async () => {
      const tx = await starknetERC20Bridge.setL2TokenBridge(BigInt(tokenBridge.address))
      await expect(tx)
        .to.emit(starknetERC20Bridge, 'LogSetL2TokenBridge')
        .withArgs(BigInt(tokenBridge.address))
    })

    it('should configure L1 token bridge MaxTotalBalance', async () => {
      await starknetERC20Bridge.setMaxTotalBalance(100000)
      const totalbalance = await starknetERC20Bridge.maxTotalBalance()
      expect(totalbalance).to.equal(100000)
    })

    it('should configure L1 token bridge MaxDeposit', async () => {
      await starknetERC20Bridge.setMaxDeposit(100)
      const deposit = await starknetERC20Bridge.maxDeposit()
      expect(deposit).to.equal(100)
    })

    it('should deposit to L1 token bridge', async () => {
      await tokenL1.approve(starknetERC20Bridge.address, 2)
      await starknetERC20Bridge.deposit(2, owner.starknetContract.address)

      const balance = await tokenL1.balanceOf(deployer.address)
      expect(balance).to.equal(8)
    })

    it('should flush L1 messages and consume on L2.', async () => {
      const flushL1Response = await starknet.devnet.flush()
      const flushL1Messages = flushL1Response.consumed_messages.from_l1
      expect(flushL1Messages).to.have.a.lengthOf(1)
      expect(flushL1Response.consumed_messages.from_l2).to.be.empty

      expect(flushL1Messages[0].args.from_address).to.equal(starknetERC20Bridge.address)
      expect(flushL1Messages[0].args.to_address).to.equal(tokenBridge.address)
      expect(flushL1Messages[0].address).to.equal(mockStarknetMessaging.address)

      const { balance } = await tokenL2.call('balanceOf', {
        account: owner.starknetContract.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(number.toBN(2))
    })
  })

  describe('withdraw (L1 <- L2)', function () {
    it('should initiate withdraw on L2 and send message to L1', async () => {
      await owner.invoke(tokenBridge, 'initiate_withdraw', {
        l1_recipient: BigInt(deployer.address),
        amount: uint256.bnToUint256(2),
      })
      let { balance } = await tokenL2.call('balanceOf', {
        account: owner.starknetContract.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(number.toBN(0))

      const flushL2Response = await starknet.devnet.flush()
      expect(flushL2Response.consumed_messages.from_l1).to.be.empty
      const flushL2Messages = flushL2Response.consumed_messages.from_l2

      expect(flushL2Messages).to.have.a.lengthOf(1)
      expect(flushL2Messages[0].from_address).to.equal(tokenBridge.address)
      // TODO: Starknet Devnet bug - 'consumed_messages.from_l2.[].to_address' not always padded to 20 bytes (expected b/c address)
      expect(hexPadStart(BigInt(flushL2Messages[0].to_address), 20 * 2)).to.equal(
        starknetERC20Bridge.address.toLowerCase(),
      )
    })

    it('should withdraw from L1 token bridge and consume the L2 message successfully', async () => {
      await starknetERC20Bridge['withdraw(uint256)'](2)

      const balance = await tokenL1.balanceOf(deployer.address)
      expect(balance).to.equal(10)
    })
  })

  after(async function () {
    manager.stop()
  })
})
