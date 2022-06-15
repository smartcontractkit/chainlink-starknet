import { starknet } from 'hardhat'
import { assert } from 'chai'
import { StarknetContract, ArgentAccount, Account } from 'hardhat/types/runtime'

const DECIMALS = 18

describe('ContractTestsFeed', function () {
  this.timeout(600_000)
  let account: Account
  let FeedContract: StarknetContract
  let ConsumerContract: StarknetContract

  before(async () => {
    account = await starknet.getAccountFromAddress(
      '0x03bf67bdc9b2ee74c7671ac1a5b36fe09b1c919bd213c3fd9d2497136be8bc0b',
      '0xb15844e15c04c8d24a4818713478eac5caff2aca4561f7f13b194acd5ada92c',
      'OpenZeppelin',
    )
    console.log('Account:', account.address, account.privateKey, account.publicKey)

    let FeedFactory = await starknet.getContractFactory('Mock_OCR2.cairo')
    FeedContract = await FeedFactory.deploy({ decimals: DECIMALS })
    console.log('FeedContract: ', FeedContract.address)

    let ConsumerFactory = await starknet.getContractFactory('OCR2_consumer.cairo')
    ConsumerContract = await ConsumerFactory.deploy({ address: FeedContract.address })
    console.log('ConsumerContract: ', ConsumerContract.address)
  })

  xit('should set store and read latest round data successfully', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await account.invoke(
        FeedContract,
        'set_latest_round_data',
        {
          transmission: { answer: 12, block_num: 1, observation_timestamp: 14325, transmission_timestamp: 87654 },
        },
        { maxFee: 30000000000000 },
      )

      await account.invoke(ConsumerContract, 'storeLatestRound', {}, { maxFee: 30000000000000 })
      const { round: round } = await ConsumerContract.call('readStoredRound', {})

      assert.equal(round.answer, 12)
      assert.equal(round.block_num, 1)
      assert.equal(round.observation_timestamp, 14325)
      assert.equal(round.transmission_timestamp, 87654)
    }
  })
  xit('should set store and read latest round data successfully for the second time', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await account.invoke(
        FeedContract,
        'set_latest_round_data',
        {
          transmission: { answer: 19, block_num: 2, observation_timestamp: 14345, transmission_timestamp: 62543 },
        },
        { maxFee: 30000000000000 },
      )

      await account.invoke(ConsumerContract, 'storeLatestRound', {}, { maxFee: 30000000000000 })
      const { round: round } = await ConsumerContract.call('readStoredRound', {})
      assert.equal(round.answer, 19)
      assert.equal(round.block_num, 2)
      assert.equal(round.observation_timestamp, 14345)
      assert.equal(round.transmission_timestamp, 62543)
    }
  })
  xit('should set store and read latest round data successfully for the third time', async () => {
    await new Promise((resolve) => setTimeout(resolve, 30000))
    {
      await account.invoke(
        FeedContract,
        'set_latest_round_data',
        {
          transmission: { answer: 42, block_num: 3, observation_timestamp: 9876, transmission_timestamp: 27839 },
        },
        { maxFee: 30000000000000 },
      )

      await account.invoke(ConsumerContract, 'storeLatestRound', {}, { maxFee: 30000000000000 })
      const { round: round } = await ConsumerContract.call('readStoredRound', {})
      assert.equal(round.answer, 42)
      assert.equal(round.block_num, 3)
      assert.equal(round.observation_timestamp, 9876)
      assert.equal(round.transmission_timestamp, 27839)
    }
  })
})
