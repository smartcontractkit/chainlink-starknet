import { starknet } from 'hardhat'
import { assert } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { account } from '@chainlink/starknet'

describe('ContractTestsMock', function () {
  this.timeout(600_000)
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let alice: Account
  let MockContract: StarknetContract
  let ConsumerContract: StarknetContract

  before(async () => {
    alice = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([{ account: alice.address, amount: 1e21 }])
    await alice.deployAccount()

    const decimals = 18
    const MockFactory = await starknet.getContractFactory('MockAggregator.cairo')
    await alice.declare(MockFactory)

    MockContract = await alice.deploy(MockFactory, { decimals: decimals })

    const ConsumerFactory = await starknet.getContractFactory('Aggregator_consumer.cairo')
    await alice.declare(ConsumerFactory)

    ConsumerContract = await alice.deploy(ConsumerFactory, {
      address: MockContract.address,
    })
  })

  it('should set and read latest round data successfully', async () => {
    await alice.invoke(MockContract, 'set_latest_round_data', {
      answer: 12,
      block_num: 1,
      observation_timestamp: 14325,
      transmission_timestamp: 87654,
    })

    const { round: round } = await ConsumerContract.call('readLatestRound', {})
    assert.equal(round.answer, 12)
    assert.equal(round.block_num, 1)
    assert.equal(round.started_at, 14325)
    assert.equal(round.updated_at, 87654)
  })

  it('should set and read latest round data successfully for the second time', async () => {
    await alice.invoke(MockContract, 'set_latest_round_data', {
      answer: 19,
      block_num: 2,
      observation_timestamp: 14345,
      transmission_timestamp: 62543,
    })

    const { round: round } = await ConsumerContract.call('readLatestRound', {})
    assert.equal(round.answer, 19)
    assert.equal(round.block_num, 2)
    assert.equal(round.started_at, 14345)
    assert.equal(round.updated_at, 62543)
  })

  it('should set and read latest round data successfully for the third time', async () => {
    await alice.invoke(MockContract, 'set_latest_round_data', {
      answer: 42,
      block_num: 3,
      observation_timestamp: 9876,
      transmission_timestamp: 27839,
    })
    const { round: round } = await ConsumerContract.call('readLatestRound', {})
    assert.equal(round.answer, 42)
    assert.equal(round.block_num, 3)
    assert.equal(round.started_at, 9876)
    assert.equal(round.updated_at, 27839)
  })

  it('should read Decimals successfully', async () => {
    const decimals = await ConsumerContract.call('readDecimals', {})
    assert.equal(decimals.decimals, 18)
  })
})
