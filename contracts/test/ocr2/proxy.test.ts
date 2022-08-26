import { assert, expect } from 'chai'
import BN from 'bn.js'
import { starknet } from 'hardhat'
import { constants, ec, encode, hash, number, uint256, stark, KeyPair } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'

interface Oracle {
  signer: KeyPair
  transmitter: Account
}

// Required to convert negative values into [0, PRIME) range
function toFelt(int: number | BigNumberish): BigNumberish {
  let prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME))
  return number.toBN(int).umod(prime)
}

describe('proxy.cairo', function () {
  this.timeout(TIMEOUT)

  let aggregatorContractFactory: StarknetContractFactory
  let proxyContractFactory: StarknetContractFactory

  let owner: Account
  let aggregator: StarknetContract
  let proxy: StarknetContract

  before(async function () {
    // assumes contract.cairo and events.cairo has been compiled
    aggregatorContractFactory = await starknet.getContractFactory('ocr2/mocks/MockAggregator')
    proxyContractFactory = await starknet.getContractFactory('ocr2/proxy')

    owner = await starknet.deployAccount('OpenZeppelin')

    aggregator = await aggregatorContractFactory.deploy({ decimals: 8 })

    proxy = await proxyContractFactory.deploy({
      owner: owner.address,
      address: aggregator.address,
    })

    console.log(proxy.address)

    // await owner.invoke(aggregator, 'set_config', config)
  })

  it('works', async () => {
    // insert round into the mock
    aggregator.invoke('set_latest_round_data', {
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
    new_aggregator.invoke('set_latest_round_data', {
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

    // query latest round, it should now point to the new aggregator
    round = (await proxy.call('latest_round_data')).round
    assert.equal(round.answer, '12')
  })
})
