import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, StarknetContractFactory, Account } from 'hardhat/types/runtime'
import { number } from 'starknet'
import { assertErrorMsg } from '../utils'

describe('SequencerUptimeFeed test', function () {
  this.timeout(300_000)

  let account: Account
  let nonOwner: Account

  // should be beforeeach, but that'd be horribly slow. Just remember that the tests are not idempotent
  before(async function () {
    account = await starknet.deployAccount('OpenZeppelin')
    nonOwner = await starknet.deployAccount('OpenZeppelin')
  })

  describe('Test inheritence', function () {
    const user = 101
    let uptimeFeedContract: StarknetContract

    before(async function () {
      const uptimeFeedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
      uptimeFeedContract = await uptimeFeedFactory.deploy({
        initial_status: 0,
        owner_address: number.toBN(account.starknetContract.address),
      })
    })

    it('Test grainting access', async function () {
      await account.invoke(uptimeFeedContract, 'add_access', {
        user,
      })

      try {
        await nonOwner.invoke(uptimeFeedContract, 'add_access', {
          user,
        })

        expect.fail()
      } catch (err: any) {
        assertErrorMsg(err?.message, 'Ownable: caller is not the owner')
      }
    })

    it('Test has_access', async function () {
      {
        const res = await uptimeFeedContract.call('has_access', { user: user, data: [] })
        expect(res.bool).to.equal(1n)
      }

      {
        const res = await uptimeFeedContract.call('has_access', { user: user + 1, data: [] })
        expect(res.bool).to.equal(0n)
      }
    })

    it('Test check_access', async function () {
      await uptimeFeedContract.call('check_access', { user: user })

      try {
        await account.invoke(uptimeFeedContract, 'check_access', { user: user + 1 })
      } catch (err: any) {
        assertErrorMsg(err?.message, 'AccessController: address does not have access')
      }
    })
  })

  describe('Test interface', function () {
    const user = 101
    let uptimeFeedContract: StarknetContract
    let mockContract: StarknetContract

    before(async function () {
      const uptimeFeedFactory = await starknet.getContractFactory('sequencer_uptime_feed')
      uptimeFeedContract = await uptimeFeedFactory.deploy({
        initial_status: 0,
        owner_address: number.toBN(account.starknetContract.address),
      })

      const mockContractFactory = await starknet.getContractFactory('uptime_feed_mock')
      mockContract = await mockContractFactory.deploy({ uptime_feed_address: number.toBN(uptimeFeedContract.address) })
    })

    it('check interface', async function () {
      {
        const res = await mockContract.call('latest_round_data')
        expect(res.round.round_id).to.equal(1n)
        expect(res.round.answer).to.equal(0n)
      }

      {
        const res = await mockContract.call('description')
        expect(res.description).to.equal(134626335741441605527772921271890603575702899782138692259993464692975953252n)
      }

      {
        const res = await mockContract.call('has_access', { user: user, data: [] })
        expect(res.bool).to.equal(0n)
      }

      await account.invoke(uptimeFeedContract, 'add_access', {
        user,
      })

      {
        const res = await mockContract.call('has_access', { user: user, data: [] })
        expect(res.bool).to.equal(1n)
      }
    })
  })
})
