import { assert, expect } from 'chai'
import { starknet } from 'hardhat'
import { ec, hash, num } from 'starknet'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'
import { account, expectInvokeError, expectSuccessOrDeclared } from '@chainlink/starknet'
import { bytesToFelts } from '@chainlink/starknet-gauntlet'

interface Oracle {
  // hex string
  signer: string
  transmitter: Account
}

// Observers - max 31 oracles or 31 bytes
const OBSERVERS_MAX = 31
const OBSERVERS_HEX = '0x00010203000000000000000000000000000000000000000000000000000000'
const UINT128_MAX = BigInt(2) ** BigInt(128) - BigInt(1)

describe('Aggregator', function () {
  this.timeout(TIMEOUT)
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let aggregatorFactory: StarknetContractFactory

  let owner: Account
  let token: StarknetContract
  let aggregator: StarknetContract

  let minAnswer = 2
  let maxAnswer = 1000000000

  let f = 1
  let n = 3 * f + 1
  let oracles: Oracle[] = []
  let config_digest: number

  let answer: string

  before(async () => {
    aggregatorFactory = await starknet.getContractFactory('aggregator')

    // can also be declared as
    // account = (await starknet.deployAccount("OpenZeppelin")) as OpenZeppelinAccount
    // if imported from hardhat/types/runtime"
    owner = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([{ account: owner.address, amount: 1e21 }])
    await owner.deployAccount()

    const tokenFactory = await starknet.getContractFactory('link_token')
    await expectSuccessOrDeclared(owner.declare(tokenFactory, { maxFee: 1e20 }))
    token = await owner.deploy(tokenFactory, {
      minter: owner.starknetContract.address,
      owner: owner.starknetContract.address,
    })

    await owner.invoke(token, 'permissionedMint', {
      account: owner.starknetContract.address,
      amount: 100_000_000_000,
    })

    await expectSuccessOrDeclared(owner.declare(aggregatorFactory, { maxFee: 1e20 }))

    aggregator = await owner.deploy(aggregatorFactory, {
      owner: BigInt(owner.starknetContract.address),
      link: BigInt(token.address),
      min_answer: minAnswer, // TODO: toFelt() to correctly wrap negative ints
      max_answer: maxAnswer, // TODO: toFelt() to correctly wrap negative ints
      billing_access_controller: 0, // TODO: billing AC
      decimals: 8,
      description: starknet.shortStringToBigInt('FOO/BAR'),
    })

    console.log(`Deployed 'aggregator.cairo': ${aggregator.address}`)

    let futures = []
    let generateOracle = async () => {
      let transmitter = await starknet.OpenZeppelinAccount.createAccount()

      await funder.fund([{ account: transmitter.address, amount: 1e21 }])
      await transmitter.deployAccount()

      return {
        signer: '0x' + Buffer.from(ec.starkCurve.utils.randomPrivateKey()).toString('hex'),
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
    let offchain_config_encoded = bytesToFelts(offchain_config)
    console.log('Encoded offchain_config: %O', offchain_config_encoded)

    let config = {
      oracles: oracles.map((oracle) => {
        return {
          signer: ec.starkCurve.getStarkKey(oracle.signer),
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

    let { response } = await aggregator.call('latest_config_details')
    config_digest = response[2]
    console.log(`Config digest: 0x${config_digest.toString(16)}`)

    // Immitate the fetch done by relay to confirm latest_config_details_works
    let block = await starknet.getBlock({ blockNumber: response.block_number })
    let events = block.transaction_receipts[0].events

    assert.isNotEmpty(events)
    assert.equal(events.length, 2)
    console.log("Log raw 'ConfigSet' event: %O", events[0])

    const decodedEvents = aggregator.decodeEvents(events)
    assert.isNotEmpty(decodedEvents)
    assert.equal(decodedEvents.length, 1)
    console.log("Log decoded 'ConfigSet' event: %O", decodedEvents[0])

    let e = decodedEvents[0]
    assert.equal(e.name, 'ConfigSet')
  })

  describe('OCR aggregator behavior', function () {
    let transmit = async (epoch_and_round: number, answer: num.BigNumberish): Promise<any> => {
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
      console.log('report data:', reportData)
      const reportDigest = hash.computeHashOnElements(reportData)
      console.log('Report data: %O', reportData)
      console.log(`Report digest: ${reportDigest}`)

      console.log('Report signatures - START')
      const signatures = []
      for (let { signer } of oracles.slice(0, f + 1)) {
        const signature = ec.starkCurve.sign(reportDigest, signer)
        const { r, s } = signature
        const starkKey = ec.starkCurve.getStarkKey(signer)
        const pubKey = '0x' + Buffer.from(ec.starkCurve.getPublicKey(signer)).toString('hex')
        signatures.push({ r, s, public_key: starkKey })
        console.log({
          starkKey,
          pubKey,
          privKey: signer,
          r,
          s,
        })
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

      const decodedEvents = aggregator.decodeEvents(receipt.events)
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

      // NOTICE: Leading zeros are trimmed for an encoded felt (number).
      //   To decode, the raw felt needs to be start padded up to max felt size (252 bits or < 32 bytes).
      const hexPadStart = (data: number | bigint, len: number) => {
        return `0x${data.toString(16).padStart(len, '0')}`
      }

      expect(hexPadStart(e.data.transmitter, len)).to.hexEqual(transmitter)

      const lenObservers = OBSERVERS_MAX * 2 // 31 bytes (hex)
      assert.equal(hexPadStart(e.data.observers, lenObservers), OBSERVERS_HEX)
      assert.equal(e.data.observations.length, 4n)

      assert.equal(hexPadStart(e.data.config_digest, len), hexPadStart(config_digest, len))
    })

    it('should transmit correctly', async () => {
      await transmit(2, 99)

      let { response: round } = await aggregator.call('latest_round_data')
      assert.equal(round.round_id, 2)
      assert.equal(round.answer, 99)

      // await transmit(3, -10) // TODO: toFelt() to correctly wrap negative ints
      // ;({ round } = await aggregator.call('latest_round_data'))
      // assert.equal(round.round_id, 3)
      // assert.equal(round.answer, -10)

      try {
        await transmit(4, 1)
        expect.fail()
      } catch (err: any) {
        // Round should be unchanged
        let { response: new_round } = await aggregator.call('latest_round_data')
        assert.deepEqual(round, new_round)
      }
    })

    it('should transmit with max u128 value correctly', async () => {
      await expectInvokeError(transmit(4, UINT128_MAX), 'median is out of min-max range')
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
      } catch (err: any) {}

      // successful transfer
      await oracle.invoke(aggregator, 'transfer_payeeship', {
        transmitter,
        proposed: proposed_payee,
      })

      // only proposed payee can accept
      try {
        await oracle.invoke(aggregator, 'accept_payeeship', { transmitter })
        expect.fail()
      } catch (err: any) {}

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
      let { response: owed } = await aggregator.call('owed_payment', {
        transmitter: oracle.transmitter.starknetContract.address,
      })
      // several rounds happened so we are owed payment
      assert.ok(owed > 0)

      const availableToValue = ([is_negative, abs_difference]: [boolean, bigint]): bigint => {
        return is_negative ? -abs_difference : abs_difference
      }

      // no funds on contract, so no LINK available for payment
      let { response: available } = await aggregator.call('link_available_for_payment')
      assert.ok(availableToValue(available) < 0) // should be negative: we owe payments

      // deposit LINK to contract
      await owner.invoke(token, 'transfer', {
        recipient: aggregator.address,
        amount: 100_000_000_000,
      })

      // we have enough funds available now
      available = (await aggregator.call('link_available_for_payment')).response
      assert.ok(availableToValue(available) > 0)

      // attempt to withdraw the payment
      await payee.invoke(aggregator, 'withdraw_payment', {
        transmitter: oracle.transmitter.starknetContract.address,
      })

      // balance as transferred to payee
      let { response: balance } = await token.call('balance_of', {
        account: payee.starknetContract.address,
      })

      assert.ok(owed === balance)

      // owed payment is now zero
      {
        let { response: owed } = await aggregator.call('owed_payment', {
          transmitter: oracle.transmitter.starknetContract.address,
        })
        assert.ok(owed == 0)
      }
    })
  })
})
