import { assert, expect } from 'chai'
import BN from 'bn.js'
import { starknet } from 'hardhat'
import { ec, hash, number, uint256, KeyPair } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { shouldBehaveLikeOwnableContract } from '../access/behavior/ownable'
import { TIMEOUT } from '../constants'
import { AccountFunder, toFelt, hexPadStart, expectInvokeErrorMsg } from '@chainlink/starknet/src/utils'

interface Oracle {
  signer: KeyPair
  transmitter: Account
}

const CHUNK_SIZE = 31

// Observers - max 31 oracles or 31 bytes
const OBSERVERS_MAX = 31
const OBSERVERS_HEX = '0x00010203000000000000000000000000000000000000000000000000000000'
const INT128_MIN = BigInt(-2) ** BigInt(128 - 1)
const INT128_MAX = BigInt(2) ** BigInt(128 - 1) - BigInt(1)

function encodeBytes(data: Uint8Array): BN[] {
  let felts = []

  // prefix with len
  let len = data.byteLength
  felts.push(number.toBN(len))

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cast to int
    felts.push(new BN(chunk, 'be'))
  }
  return felts
}

function decodeBytes(felts: BN[]): Uint8Array {
  let data = []

  // TODO: validate len > 1

  // TODO: validate it fits into 54 bits
  let length = felts.shift()?.toNumber()!

  for (const felt of felts) {
    let chunk = felt.toArray('be', Math.min(CHUNK_SIZE, length))
    data.push(...chunk)

    length -= chunk.length
  }

  return new Uint8Array(data)
}

describe('aggregator.cairo', function () {
  this.timeout(TIMEOUT)

  let aggregatorFactory: StarknetContractFactory

  let owner: Account
  let token: StarknetContract
  let aggregator: StarknetContract
  let funder: AccountFunder

  let minAnswer = -10
  let maxAnswer = 1000000000

  let f = 1
  let n = 3 * f + 1
  let oracles: Oracle[] = []
  let config_digest: number

  let answer: string

  before(async () => {
    // assumes contract.cairo and events.cairo has been compiled
    aggregatorFactory = await starknet.getContractFactory('ocr2/aggregator')

    // can also be declared as
    // account = (await starknet.deployAccount("OpenZeppelin")) as OpenZeppelinAccount
    // if imported from hardhat/types/runtime"
    owner = await starknet.deployAccount('OpenZeppelin')
    const opts = { network: 'devnet' }
    funder = new AccountFunder(opts)
    await funder.fund([{ account: owner.address, amount: 5000 }])

    const tokenFactory = await starknet.getContractFactory('link_token')
    token = await tokenFactory.deploy({ owner: owner.starknetContract.address })

    await owner.invoke(token, 'permissionedMint', {
      account: owner.starknetContract.address,
      amount: uint256.bnToUint256(100_000_000_000),
    })

    aggregator = await aggregatorFactory.deploy({
      owner: BigInt(owner.starknetContract.address),
      link: BigInt(token.address),
      min_answer: toFelt(minAnswer),
      max_answer: toFelt(maxAnswer),
      billing_access_controller: 0, // TODO: billing AC
      decimals: 8,
      description: starknet.shortStringToBigInt('FOO/BAR'),
    })

    console.log(`Deployed 'aggregator.cairo': ${aggregator.address}`)

    let futures = []
    let generateOracle = async () => {
      let transmitter = await starknet.deployAccount('OpenZeppelin')
      await funder.fund([{ account: owner.address, amount: 5000 }])
      return {
        signer: ec.genKeyPair(),
        transmitter,
        // payee
      }
    }
    for (let i = 0; i < n; i++) {
      futures.push(generateOracle())
    }
    oracles = await Promise.all(futures)

    let onchain_config: number[] = []
    let offchain_config_version = 2
    let offchain_config = new Uint8Array([1])
    let offchain_config_encoded = encodeBytes(offchain_config)
    console.log('Encoded offchain_config: %O', encodeBytes(offchain_config))

    let config = {
      oracles: oracles.map((oracle) => {
        return {
          signer: number.toBN(ec.getStarkKey(oracle.signer)),
          transmitter: oracle.transmitter.starknetContract.address,
        }
      }),
      f,
      onchain_config,
      offchain_config_version,
      offchain_config: offchain_config_encoded,
    }
    await owner.invoke(aggregator, 'set_config', config)
    console.log('Config: %O', config)

    let result = await aggregator.call('latest_config_details')
    config_digest = result.config_digest
    console.log(`Config digest: 0x${config_digest.toString(16)}`)

    // Immitate the fetch done by relay to confirm latest_config_details_works
    let block = await starknet.getBlock({ blockNumber: result.block_number })
    let events = block.transaction_receipts[0].events

    assert.isNotEmpty(events)
    assert.equal(events.length, 1)
    console.log("Log raw 'ConfigSet' event: %O", events[0])

    const decodedEvents = await aggregator.decodeEvents(events)
    assert.isNotEmpty(decodedEvents)
    assert.equal(decodedEvents.length, 1)
    console.log("Log decoded 'ConfigSet' event: %O", decodedEvents[0])

    let e = decodedEvents[0]
    assert.equal(e.name, 'ConfigSet')
  })

  shouldBehaveLikeOwnableContract(async () => {
    const alice = owner
    const bob = await starknet.deployAccount('OpenZeppelin')
    await funder.fund([{ account: bob.address, amount: 5000 }])
    return { ownable: aggregator, alice, bob }
  })

  describe('OCR aggregator behavior', function () {
    let transmit = async (epoch_and_round: number, answer: BigNumberish): Promise<any> => {
      let extra_hash = 1
      let observation_timestamp = 1
      let juels_per_fee_coin = 1
      let gas_price = 1

      let observers_buf = Buffer.alloc(31)
      let observations = []

      for (const [index, _] of oracles.entries()) {
        observers_buf[index] = index
        observations.push(answer)
      }

      // convert to a single value that will be decoded by toBN
      let observers = `0x${observers_buf.toString('hex')}`
      assert.equal(observers, OBSERVERS_HEX)

      const reportData = [
        // report_context
        config_digest,
        epoch_and_round,
        extra_hash,
        // raw_report
        observation_timestamp,
        observers,
        observations.length,
        ...observations,
        juels_per_fee_coin,
        gas_price,
      ]
      let reportDigest = hash.computeHashOnElements(reportData)
      console.log('Report data: %O', reportData)
      console.log(`Report digest: ${reportDigest}`)

      console.log('Report signatures - START')
      let signatures = []
      for (let { signer } of oracles.slice(0, f + 1)) {
        let [r, s] = ec.sign(signer, reportDigest)
        const privKey = signer.getPrivate()
        const pubKey = number.toBN(ec.getStarkKey(signer))
        signatures.push({ r, s, public_key: pubKey })
        console.log({ pubKey, privKey, r, s })
      }
      console.log('Report signatures - END\n')

      const transmitter = oracles[0].transmitter
      return await transmitter.invoke(aggregator, 'transmit', {
        report_context: {
          config_digest,
          epoch_and_round,
          extra_hash,
        },
        observation_timestamp,
        observers,
        observations,
        juels_per_fee_coin,
        gas_price,
        signatures,
      })
    }

    it("should 'set_billing' successfully", async () => {
      await owner.invoke(aggregator, 'set_billing', {
        config: {
          observation_payment_gjuels: 1,
          transmission_payment_gjuels: 5,
          gas_base: 1,
          gas_per_signature: 1,
        },
      })
    })

    it("should emit 'NewTransmission' event on transmit", async () => {
      const txHash = await transmit(1, 99)
      const receipt = await starknet.getTransactionReceipt(txHash)

      assert.isNotEmpty(receipt.events)
      console.log("Log raw 'NewTransmission' event: %O", receipt.events[0])

      const decodedEvents = await aggregator.decodeEvents(receipt.events)
      assert.isNotEmpty(decodedEvents)
      console.log("Log decoded 'NewTransmission' event: %O", decodedEvents[0])

      const e = decodedEvents[0]
      const transmitter = oracles[0].transmitter.address

      assert.equal(e.name, 'NewTransmission')
      assert.equal(e.data.round_id, 1n)
      assert.equal(e.data.observation_timestamp, 1n)
      assert.equal(e.data.epoch_and_round, 1n)
      // assert.equal(e.data.reimbursement, 0n)

      const len = 32 * 2 // 32 bytes (hex)
      assert.equal(hexPadStart(e.data.transmitter, len), transmitter)

      const lenObservers = OBSERVERS_MAX * 2 // 31 bytes (hex)
      assert.equal(hexPadStart(e.data.observers, lenObservers), OBSERVERS_HEX)
      assert.equal(e.data.observations_len, 4n)

      assert.equal(hexPadStart(e.data.config_digest, len), hexPadStart(config_digest, len))
    })

    it('should transmit correctly', async () => {
      await transmit(2, 99)

      let { round } = await aggregator.call('latest_round_data')
      assert.equal(round.round_id, 2)
      assert.equal(round.answer, 99)

      await transmit(3, toFelt(-10))
        ; ({ round } = await aggregator.call('latest_round_data'))
      assert.equal(round.round_id, 3)
      assert.equal(round.answer, -10)

      try {
        await transmit(4, -100)
        expect.fail()
      } catch (err: any) {
        // Round should be unchanged
        let { round: new_round } = await aggregator.call('latest_round_data')
        assert.deepEqual(round, new_round)
      }
    })

    it('should transmit with max_int_128bit correctly', async () => {
      answer = BigInt(INT128_MAX).toString(10)
      try {
        await transmit(4, toFelt(answer))
        expect.fail()
      } catch (error: any) {
        const matches = error?.message.match(/Error message: (.+?)\n/g)
        if (!matches) {
          console.log('answer is in int128 range but not in min-max range')
        }
      }

      try {
        const tooBig = INT128_MAX + 1n
        answer = BigInt(tooBig).toString(10)
        await transmit(4, answer)
        expect.fail()
      } catch (err: any) {
        expectInvokeErrorMsg(
          err?.message,
          `Error message: Aggregator: value not in int128 range: ${answer}\n`,
        )
      }
    })

    it('should transmit with min_int_128bit correctly', async () => {
      answer = BigInt(INT128_MIN).toString(10)
      try {
        await transmit(4, toFelt(answer))
      } catch (err: any) {
        const matches = err?.message.match(/Error message: (.+?)\n/g)
        if (!matches) {
          console.log('answer is in int128 range but not in min-max range')
        }
      }

      try {
        const tooBig = INT128_MIN - 1n
        answer = BigInt(tooBig).toString(10)
        await transmit(4, toFelt(answer))
        expect.fail()
      } catch (err: any) {
        expectInvokeErrorMsg(
          err?.message,
          `Error message: Aggregator: value not in int128 range: ${answer}\n`,
        )
      }
    })

    it('payee management', async () => {
      let payees = oracles.map((oracle) => ({
        transmitter: oracle.transmitter.starknetContract.address,
        payee: oracle.transmitter.starknetContract.address, // reusing transmitter acocunts as payees for simplicity
      }))
      // call set_payees, should succeed because all payees are zero
      await owner.invoke(aggregator, 'set_payees', { payees })
      // call set_payees, should succeed because values are unchanged
      await owner.invoke(aggregator, 'set_payees', { payees })

      let oracle = oracles[0].transmitter
      let transmitter = oracle.starknetContract.address
      let payee = transmitter

      let proposed_oracle = oracles[1].transmitter
      let proposed_transmitter = proposed_oracle.starknetContract.address
      let proposed_payee = proposed_transmitter

      // can't transfer to self
      try {
        await oracle.invoke(aggregator, 'transfer_payeeship', {
          transmitter,
          proposed: payee,
        })
        expect.fail()
      } catch (err: any) {
        // TODO: expect(err.message).to.contain("");
      }

      // only payee can transfer
      try {
        await proposed_oracle.invoke(aggregator, 'transfer_payeeship', {
          transmitter,
          proposed: proposed_payee,
        })
        expect.fail()
      } catch (err: any) { }

      // successful transfer
      await oracle.invoke(aggregator, 'transfer_payeeship', {
        transmitter,
        proposed: proposed_payee,
      })

      // only proposed payee can accept
      try {
        await oracle.invoke(aggregator, 'accept_payeeship', { transmitter })
        expect.fail()
      } catch (err: any) { }

      // successful accept
      await proposed_oracle.invoke(aggregator, 'accept_payeeship', {
        transmitter,
      })
    })

    it('payments and withdrawals', async () => {
      let oracle = oracles[0]
      // NOTE: previous test changed oracle0's payee to oracle1
      let payee = oracles[1].transmitter
      aggregator.call
      let { amount: owed } = await aggregator.call('owed_payment', {
        transmitter: oracle.transmitter.starknetContract.address,
      })
      // several rounds happened so we are owed payment
      assert.ok(owed > 0)

      // no funds on contract, so no LINK available for payment
      let { available } = await aggregator.call('link_available_for_payment')
      assert.ok(available < 0) // should be negative: we owe payments

      // deposit LINK to contract
      await owner.invoke(token, 'transfer', {
        recipient: aggregator.address,
        amount: uint256.bnToUint256(100_000_000_000),
      })

      // we have enough funds available now
      available = (await aggregator.call('link_available_for_payment')).available
      assert.ok(available > 0)

      // attempt to withdraw the payment
      await payee.invoke(aggregator, 'withdraw_payment', {
        transmitter: oracle.transmitter.starknetContract.address,
      })

      // balance as transferred to payee
      let { balance } = await token.call('balanceOf', {
        account: payee.starknetContract.address,
      })

      assert.ok(number.toBN(owed).eq(uint256.uint256ToBN(balance)))

      // owed payment is now zero
      {
        let { amount: owed } = await aggregator.call('owed_payment', {
          transmitter: oracle.transmitter.starknetContract.address,
        })
        assert.ok(owed == 0)
      }
    })
  })
})
