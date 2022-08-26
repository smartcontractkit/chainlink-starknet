import { starknet } from 'hardhat'
import { expect } from 'chai'
import { StarknetContract, ArgentAccount } from 'hardhat/types/runtime'
import { TIMEOUT } from './constants'

const NAME = starknet.shortStringToBigInt('Chainlink')
const SYMBOL = starknet.shortStringToBigInt('LINK')
const DECIMALS = 18
const L1_BRIDGE_ADDRESS = 42n

function adaptAddress(address: string) {
  return '0x' + BigInt(address).toString(16)
}

function expectAddressEquality(actual: string, expected: string) {
  expect(adaptAddress(actual)).to.equal(adaptAddress(expected))
}

describe('ContractTokenBridge', function () {
  this.timeout(TIMEOUT)
  let accountUser1: ArgentAccount
  let tokenBridgeContract: StarknetContract
  let ERC20Contract: StarknetContract

  before(async () => {
    accountUser1 = (await starknet.deployAccount('Argent')) as ArgentAccount

    let tokenBridgeFactory = await starknet.getContractFactory('token_bridge.cairo')
    tokenBridgeContract = await tokenBridgeFactory.deploy({
      governor_address: accountUser1.starknetContract.address,
    })

    let ERC20Factory = await starknet.getContractFactory('ERC20.cairo')
    ERC20Contract = await ERC20Factory.deploy({
      name: NAME,
      symbol: SYMBOL,
      decimals: DECIMALS,
      minter_address: tokenBridgeContract.address,
    })
  })

  it('Test Set and Get function for L1 bridge address', async () => {
    await accountUser1.invoke(tokenBridgeContract, 'set_l1_bridge', {
      l1_bridge_address: L1_BRIDGE_ADDRESS,
    })
    const { res: l1_address } = await tokenBridgeContract.call('get_l1_bridge', {})
    expectAddressEquality(l1_address.toString(), L1_BRIDGE_ADDRESS.toString())
  })

  it('Test Set and Get function for L2 token address', async () => {
    await accountUser1.invoke(tokenBridgeContract, 'set_l2_token', {
      l2_token_address: ERC20Contract.address,
    })
    const { res: l2_address } = await tokenBridgeContract.call('get_l2_token', {})
    expectAddressEquality(l2_address.toString(), ERC20Contract.address)
  })
})
