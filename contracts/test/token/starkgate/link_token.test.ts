import { account } from '@chainlink/starknet'
import { starknet } from 'hardhat'
import { TIMEOUT } from '../../constants'
import { shouldBehaveLikeStarkGateERC20 } from './behavior/ERC20'

describe('link_token', function () {
  this.timeout(TIMEOUT)
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  shouldBehaveLikeStarkGateERC20(async () => {
    const owner = await starknet.OpenZeppelinAccount.createAccount()
    const alice = await starknet.OpenZeppelinAccount.createAccount()
    const bob = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([
      { account: owner.address, amount: 1e21 },
      { account: alice.address, amount: 1e21 },
      { account: bob.address, amount: 1e21 },
    ])
    await owner.deployAccount()
    await alice.deployAccount()
    await bob.deployAccount()

    const tokenFactory = await starknet.getContractFactory('link_token')
    await owner.declare(tokenFactory)
    const token = await owner.deploy(tokenFactory, { owner: owner.starknetContract.address })
    return { token, owner, alice, bob }
  })
})
