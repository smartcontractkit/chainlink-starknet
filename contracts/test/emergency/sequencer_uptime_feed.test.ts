import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { number } from 'starknet'
import { shouldBehaveLikeOwnableContract } from '../access/behavior/ownable'
import { account, expectInvokeError } from '@chainlink/starknet'

describe('SequencerUptimeFeed', function () {
  this.timeout(300_000)

  let owner: Account
  let nonOwner: Account
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)
  // should be beforeeach, but that'd be horribly slow. Just remember that the tests are not idempotent
  before(async function () {
    owner = await starknet.OpenZeppelinAccount.createAccount()
    nonOwner = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([
      { account: owner.address, amount: 9000000000000000 },
      { account: nonOwner.address, amount: 9000000000000000 },
    ])
    await owner.deployAccount()
    await nonOwner.deployAccount()
  })

  shouldBehaveLikeOwnableContract(async () => {
    const alice = owner
    const bob = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([{ account: bob.address, amount: 9000000000000000 }])

    await bob.deployAccount()

    const feedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
    await alice.declare(feedFactory)

    const feed = await alice.deploy(feedFactory, {
      initial_status: 0,
      owner_address: number.toBN(alice.starknetContract.address),
    })

    return { ownable: feed, alice, bob }
  })

  describe('Test access control via inherited `SimpleReadAccessController`', function () {
    const user = 101
    let uptimeFeedContract: StarknetContract

    before(async function () {
      const uptimeFeedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
      await owner.declare(uptimeFeedFactory)

      uptimeFeedContract = await owner.deploy(uptimeFeedFactory, {
        initial_status: 0,
        owner_address: number.toBN(owner.starknetContract.address),
      })
    })

    it('should block non-owners from making admin changes', async function () {
      await owner.invoke(uptimeFeedContract, 'add_access', { user })

      await expectInvokeError(
        nonOwner.invoke(uptimeFeedContract, 'add_access', { user }),
        'Ownable: caller is not the owner',
      )
    })

    it('should report access information correctly', async function () {
      {
        const res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
        expect(res.bool).to.equal(1n)
      }

      {
        const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
        expect(res.bool).to.equal(0n)
      }
    })

    it('should error on `check_access` without access', async function () {
      await uptimeFeedContract.call('check_access', { user: user })

      await expectInvokeError(
        owner.invoke(uptimeFeedContract, 'check_access', { user: user + 1 }),
        'SimpleReadAccessController: address does not have access',
      )
    })

    it('should disable access check', async function () {
      await owner.invoke(uptimeFeedContract, 'disable_access_check', {})

      const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
      expect(res.bool).to.equal(1n)
    })

    it('should enable access check', async function () {
      await owner.invoke(uptimeFeedContract, 'enable_access_check', {})

      const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
      expect(res.bool).to.equal(0n)
    })

    it('should remove user access', async function () {
      const res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
      expect(res.bool).to.equal(1n)

      await owner.invoke(uptimeFeedContract, 'remove_access', { user: user })

      const new_res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
      expect(new_res.bool).to.equal(0n)
    })
  })

  describe('Test IAggregator interface using a Proxy', function () {
    let uptimeFeedContract: StarknetContract
    let proxyContract: StarknetContract

    before(async function () {
      const uptimeFeedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
      await owner.declare(uptimeFeedFactory)
      uptimeFeedContract = await owner.deploy(uptimeFeedFactory, {
        initial_status: 0,
        owner_address: number.toBN(owner.starknetContract.address),
      })

      const proxyFactory = await starknet.getContractFactory('aggregator_proxy')
      await owner.declare(proxyFactory)
      proxyContract = await owner.deploy(proxyFactory, {
        owner: number.toBN(owner.starknetContract.address),
        address: number.toBN(uptimeFeedContract.address),
      })

      // proxy contract needs to have access to uptimeFeedContract
      await owner.invoke(uptimeFeedContract, 'add_access', { user: proxyContract.address })
    })

    it('should block access when using an account without access', async function () {
      const accWithoutAccess = await starknet.OpenZeppelinAccount.createAccount()

      await funder.fund([{ account: accWithoutAccess.address, amount: 9000000000000000 }])
      await accWithoutAccess.deployAccount()
      await expectInvokeError(
        accWithoutAccess.invoke(proxyContract, 'latest_round_data'),
        'SimpleReadAccessController: address does not have access',
      )
    })

    it('should respond via an aggregator_proxy contract', async function () {
      {
        const res = await proxyContract.call('latest_round_data')
        expect(res.round.answer).to.equal(0n)
      }

      {
        const res = await proxyContract.call('description')
        expect(res.description).to.equal(
          134626335741441605527772921271890603575702899782138692259993464692975953252n,
        )
      }

      {
        const res = await proxyContract.call('decimals')
        expect(res.decimals).to.equal(0n)
      }

      // TODO: enable access check and assert correct behaviour
    })
  })
})
