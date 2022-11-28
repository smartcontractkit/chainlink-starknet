import { assert } from 'chai'
import { starknet } from 'hardhat'
import { number } from 'starknet'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'
import { shouldBehaveLikeOwnableContract } from '../access/behavior/ownable'
import { account, loadConfig, NetworkManager, FunderOptions, Funder } from '@chainlink/starknet'

describe('aggregator_proxy.cairo', function () {
  this.timeout(TIMEOUT)

  const config = loadConfig()
  const optsConf = { config, required: ['devnet'] }
  const manager = new NetworkManager(optsConf)

  let aggregatorContractFactory: StarknetContractFactory
  let proxyContractFactory: StarknetContractFactory

  let owner: Account
  let aggregator: StarknetContract
  let proxy: StarknetContract
  let opts: FunderOptions
  let funder: Funder

  before(async function () {
    await manager.start()
    opts = account.makeFunderOptsFromEnv()
    funder = new account.Funder(opts)
    // assumes contract.cairo and events.cairo has been compiled
    aggregatorContractFactory = await starknet.getContractFactory('ocr2/mocks/MockAggregator')
    proxyContractFactory = await starknet.getContractFactory('ocr2/aggregator_proxy')

    owner = await starknet.deployAccount('OpenZeppelin')
    await funder.fund([{ account: owner.address, amount: 5000 }])

    aggregator = await aggregatorContractFactory.deploy({ decimals: 8 })

    proxy = await proxyContractFactory.deploy({
      owner: owner.address,
      address: aggregator.address,
    })

    console.log(proxy.address)
  })

  shouldBehaveLikeOwnableContract(async () => {
    const alice = owner
    const bob = await starknet.deployAccount('OpenZeppelin')
    await funder.fund([{ account: bob.address, amount: 5000 }])

    return { ownable: proxy, alice, bob }
  })

  describe('proxy behaviour', function () {
    it('works', async () => {
      // insert round into the mock
      await owner.invoke(aggregator, 'set_latest_round_data', {
        answer: 10,
        block_num: 1,
        observation_timestamp: 9,
        transmission_timestamp: 8,
      })

      // query latest round
      let { round } = await proxy.call('latest_round_data')
      // TODO: split_felt the round_id and check phase=1 round=1
      assert.equal(round.answer, '10')
      assert.equal(round.block_num, '1')
      assert.equal(round.started_at, '9')
      assert.equal(round.updated_at, '8')

      // insert a second ocr2 aggregator
      let new_aggregator = await aggregatorContractFactory.deploy({ decimals: 8 })

      // insert round into the mock
      await owner.invoke(new_aggregator, 'set_latest_round_data', {
        answer: 12,
        block_num: 2,
        observation_timestamp: 10,
        transmission_timestamp: 11,
      })

      // propose it to the proxy
      await owner.invoke(proxy, 'propose_aggregator', {
        address: new_aggregator.address,
      })

      // query latest round, it should still point to the old aggregator
      round = (await proxy.call('latest_round_data')).round
      assert.equal(round.answer, '10')

      // but the proposed round should be newer
      round = (await proxy.call('proposed_latest_round_data')).round
      assert.equal(round.answer, '12')

      // confirm the new aggregator
      await owner.invoke(proxy, 'confirm_aggregator', {
        address: new_aggregator.address,
      })

      const phase_aggregator = await proxy.call('aggregator', {})
      assert.equal(phase_aggregator.aggregator, number.toBN(new_aggregator.address))

      const phase_id = await proxy.call('phase_id', {})
      assert.equal(phase_id.phase_id, 2n)

      // query latest round, it should now point to the new aggregator
      round = (await proxy.call('latest_round_data')).round
      assert.equal(round.answer, '12')
    })
  })

  after(async function () {
    manager.stop()
  })
})
