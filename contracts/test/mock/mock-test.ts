import { starknet } from 'hardhat'
import { assert } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'

const DECIMALS = 18

describe('ContractTestsMock', function () {
  this.timeout(600_000)
  let account: Account
  let MockContract: StarknetContract
  let ConsumerContract: StarknetContract

  before(async () => {
    account = await starknet.deployAccount('OpenZeppelin')

    let MockFactory = await starknet.getContractFactory('examples/contracts/Mock_Aggregator')
    MockContract = await MockFactory.deploy({ decimals: DECIMALS })

    let ConsumerFactory = await starknet.getContractFactory('examples/contracts/OCR2_consumer')
    ConsumerContract = await ConsumerFactory.deploy({ address: MockContract.address })
  })

  it('should set and read latest round data successfully', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await MockContract.invoke('set_latest_round_data', {
        answer: 12,
        block_num: 1,
        observation_timestamp: 14325,
        transmission_timestamp: 87654,
      })

      const { round: round } = await ConsumerContract.call('readLatestRound', {})
      assert.equal(round.answer, 12)
      assert.equal(round.block_num, 1)
      assert.equal(round.observation_timestamp, 14325)
      assert.equal(round.transmission_timestamp, 87654)
    }
  })
  it('should set and read latest round data successfully for the second time', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await MockContract.invoke('set_latest_round_data', {
        answer: 19,
        block_num: 2,
        observation_timestamp: 14345,
        transmission_timestamp: 62543,
      })
      const { round: round } = await ConsumerContract.call('readLatestRound', {})
      assert.equal(round.answer, 19)
      assert.equal(round.block_num, 2)
      assert.equal(round.observation_timestamp, 14345)
      assert.equal(round.transmission_timestamp, 62543)
    }
  })
  it('should set and read latest round data successfully for the third time', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await MockContract.invoke('set_latest_round_data', {
        answer: 42,
        block_num: 3,
        observation_timestamp: 9876,
        transmission_timestamp: 27839,
      })
      const { round: round } = await ConsumerContract.call('readLatestRound', {})
      assert.equal(round.answer, 42)
      assert.equal(round.block_num, 3)
      assert.equal(round.observation_timestamp, 9876)
      assert.equal(round.transmission_timestamp, 27839)
    }
  })

  it('should read Decimals successfully', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      const decimals = await ConsumerContract.call('readDecimals', {})
      console.log(decimals)
      assert.equal(decimals.decimals, 18)
    }
  })
})
