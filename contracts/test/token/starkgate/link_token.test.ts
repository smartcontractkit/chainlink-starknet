import { starknet } from 'hardhat'
import { TIMEOUT } from '../../constants'
import { shouldBehaveLikeStarkGateERC20 } from '../starkgate/behavior/ERC20'

describe('link_token', function () {
  this.timeout(TIMEOUT)

  shouldBehaveLikeStarkGateERC20(async () => {
    const owner = await starknet.deployAccount('OpenZeppelin')

    const tokenFactory = await starknet.getContractFactory('link_token')
    const token = await tokenFactory.deploy({ owner: owner.starknetContract.address })

    const alice = await starknet.deployAccount('OpenZeppelin')
    const bob = await starknet.deployAccount('OpenZeppelin')
    return { token, owner, alice, bob }
  })
})
