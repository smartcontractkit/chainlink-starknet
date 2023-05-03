import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account, StarknetContractFactory } from 'hardhat/types/runtime'
import { num } from 'starknet'
import { shouldBehaveLikeOwnableContract } from '../access/behavior/ownable'
import { account, expectInvokeError } from '@chainlink/starknet'

describe('SequencerUptimeFeed', function () {
  this.timeout(1_000_000)

  let owner: Account
  let nonOwner: Account
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)
  let feedFactory: StarknetContractFactory
  let proxyFactory: StarknetContractFactory

  // should be beforeeach, but that'd be horribly slow. Just remember that the tests are not idempotent
  before(async function () {
    owner = await starknet.OpenZeppelinAccount.createAccount()
    nonOwner = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([
      { account: owner.address, amount: 1e21 },
      { account: nonOwner.address, amount: 1e21 },
    ])
    await owner.deployAccount()
    await nonOwner.deployAccount()

    feedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    owner.declare(feedFactory, { maxFee: 1e20 })
    proxyFactory = await starknet.getContractFactory('aggregator_proxy')
    owner.declare(proxyFactory, { maxFee: 1e20 })
  })

  shouldBehaveLikeOwnableContract(async () => {
    const alice = owner
    const bob = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([{ account: bob.address, amount: 1e21 }])

    await bob.deployAccount()

    const feed = await alice.deploy(
      feedFactory,
      {
        initial_status: 0,
        owner_address: num.toBigInt(alice.starknetContract.address),
      },
      { maxFee: 1e20 },
    )

    return { ownable: feed, alice, bob }
  })

  describe('Test access control via inherited `SimpleReadAccessController`', function () {
    const user = 101
    let uptimeFeedContract: StarknetContract

    before(async function () {
      uptimeFeedContract = await owner.deploy(feedFactory, {
        initial_status: 0,
        owner_address: num.toBigInt(owner.starknetContract.address),
      })
    })

    it('should block non-owners from making admin changes', async function () {
      await owner.invoke(uptimeFeedContract, 'add_access', { user })

      await expectInvokeError(
        nonOwner.invoke(uptimeFeedContract, 'add_access', { user }),
        'Ownable: caller is not owner',
      )
    })

    it('should report access information correctly', async function () {
      {
        const res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
        expect(res.response).to.equal(true)
      }

      {
        const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
        expect(res.response).to.equal(false)
      }
    })

    it('should error on `check_access` without access', async function () {
      await uptimeFeedContract.call('check_access', { user: user })

      await expectInvokeError(
        owner.invoke(uptimeFeedContract, 'check_access', { user: user + 1 }),
        'address does not have access',
      )
    })

    it('should disable access check', async function () {
      await owner.invoke(uptimeFeedContract, 'disable_access_check', {})

      const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
      expect(res.response).to.equal(true)
    })

    it('should enable access check', async function () {
      await owner.invoke(uptimeFeedContract, 'enable_access_check', {})

      const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
      expect(res.response).to.equal(false)
    })

    it('should remove user access', async function () {
      const res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
      expect(res.response).to.equal(true)

      await owner.invoke(uptimeFeedContract, 'remove_access', { user: user })

      const new_res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
      expect(new_res.response).to.equal(false)
    })
  })

  describe('Test IAggregator interface using a Proxy', function () {
    let uptimeFeedContract: StarknetContract
    let proxyContract: StarknetContract

    before(async function () {
      uptimeFeedContract = await owner.deploy(feedFactory, {
        initial_status: 0,
        owner_address: num.toBigInt(owner.starknetContract.address),
      })

      proxyContract = await owner.deploy(proxyFactory, {
        owner: num.toBigInt(owner.starknetContract.address),
        address: num.toBigInt(uptimeFeedContract.address),
      })

      // proxy contract needs to have access to uptimeFeedContract
      await owner.invoke(uptimeFeedContract, 'add_access', { user: proxyContract.address })
    })

    it('should block access when using an account without access', async function () {
      const accWithoutAccess = await starknet.OpenZeppelinAccount.createAccount()

      await funder.fund([{ account: accWithoutAccess.address, amount: 1e21 }])
      await accWithoutAccess.deployAccount()
      await expectInvokeError(
        accWithoutAccess.invoke(proxyContract, 'latest_round_data'),
        'address does not have access',
      )
    })

    it('should respond via an aggregator_proxy contract', async function () {
      {
        const res = await proxyContract.call('latest_round_data')
        expect(res.response.answer).to.equal(0n)
      }

      {
        const res = await proxyContract.call('description')
        expect(res.response).to.equal(
          134626335741441605527772921271890603575702899782138692259993464692975953252n,
        )
      }

      {
        const res = await proxyContract.call('decimals')
        expect(res.response).to.equal(0n)
      }

      // TODO: enable access check and assert correct behaviour
    })
  })
})
