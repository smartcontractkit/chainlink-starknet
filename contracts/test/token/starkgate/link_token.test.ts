import { AccountFunder } from '@chainlink/starknet/src/utils'
import { starknet } from 'hardhat'
import { TIMEOUT } from '../../constants'
import { shouldBehaveLikeStarkGateERC20 } from './behavior/ERC20'

describe('link_token', function () {
  this.timeout(TIMEOUT)

  shouldBehaveLikeStarkGateERC20(async () => {
    const owner = await starknet.deployAccount('OpenZeppelin')

    const tokenFactory = await starknet.getContractFactory('link_token')
    const token = await tokenFactory.deploy({ owner: owner.starknetContract.address })

    const alice = await starknet.deployAccount('OpenZeppelin')
    const bob = await starknet.deployAccount('OpenZeppelin')
    const opts = { network: 'devnet' }
    const funder = new AccountFunder(opts)
    await funder.fund([
      { account: owner.address, amount: 5000 },
      { account: alice.address, amount: 5000 },
      { account: bob.address, amount: 5000 },
    ])
    return { token, owner, alice, bob }
  })
})
