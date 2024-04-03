import { fetchStarknetAccount, getStarknetContractArtifacts } from '../utils'
import { bytesToFelts } from '@chainlink/starknet-gauntlet'
import { STARKNET_DEVNET_URL, TIMEOUT } from '../constants'
import * as account from '../account'
import { assert, expect } from 'chai'
import {
  BigNumberish,
  ParsedStruct,
  LibraryError,
  RpcProvider,
  Contract,
  CallData,
  Account,
  Uint256,
  cairo,
  hash,
  num,
  ec,
} from 'starknet'

type Oracle = Readonly<{
  // hex string
  signer: string
  transmitter: Account
}>

// Observers - max 31 oracles or 31 bytes
const OBSERVERS_MAX = 31
const OBSERVERS_HEX = '0x00010203000000000000000000000000000000000000000000000000000000'
const UINT128_MAX = BigInt(2) ** BigInt(128) - BigInt(1)

describe('Aggregator', function () {
  this.timeout(TIMEOUT)
  const provider = new RpcProvider({ nodeUrl: STARKNET_DEVNET_URL })
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let aggregator: Contract
  let token: Contract
  let owner: Account

  const maxAnswer = 1000000000
  const minAnswer = 2
  const f = 1
  const n = 3 * f + 1
  const oracles: Oracle[] = []
  let config_digest: string

  before(async () => {
    // Sets up the owner account
    owner = await fetchStarknetAccount()
    await funder.fund([{ account: owner.address, amount: 1e21 }])
    console.log('Owner account has been funded')

    // Declares and deploys the LINK token contract
    const ddToken = await owner.declareAndDeploy({
      ...getStarknetContractArtifacts('LinkToken'),
      constructorCalldata: CallData.compile({
        minter: owner.address,
        owner: owner.address,
      }),
    })
    console.log(`Successfully deployed LinkToken: ${ddToken.deploy.address}`)

    // Creates a starknet contract instance for token
    const { abi: tokenAbi } = await provider.getClassByHash(ddToken.declare.class_hash)
    token = new Contract(tokenAbi, ddToken.deploy.address, provider)

    // Funds the owner account with some LINK
    await owner.execute(
      token.populate('permissioned_mint', {
        account: owner.address,
        amount: cairo.uint256(100_000_000_000n),
      }),
    )
    console.log('Successfully funded owner account with LINK')

    // Performs the following in parallel:
    //   Deploys the aggregator contract
    //   Populates the oracles array with devnet accounts
    const [ddAggregator] = await Promise.all([
      // Declares and deploys the aggregator
      owner.declareAndDeploy({
        ...getStarknetContractArtifacts('Aggregator'),
        constructorCalldata: CallData.compile({
          owner: owner.address,
          link: token.address,
          min_answer: minAnswer, // TODO: toFelt() to correctly wrap negative ints
          max_answer: maxAnswer, // TODO: toFelt() to correctly wrap negative ints
          billing_access_controller: 0, // TODO: billing AC
          decimals: 8,
          description: 0,
        }),
      }),

      // Populates the oracles array with devnet accounts
      ...Array.from({ length: n }).map(async (_, i) => {
        // account index 0 is taken by the owner account, so we need to offset by 1
        const transmitter = await fetchStarknetAccount({ accountIndex: i + 1 })
        await funder.fund([{ account: transmitter.address, amount: 1e21 }])
        oracles.push({
          signer: '0x' + Buffer.from(ec.starkCurve.utils.randomPrivateKey()).toString('hex'),
          transmitter,
          // payee
        })
      }),
    ])
    console.log(`Successfully deployed Aggregator: ${ddAggregator.deploy.address}`)

    // Creates a starknet contract instance for aggregator
    const { abi: aggregatorAbi } = await provider.getClassByHash(ddAggregator.declare.class_hash)
    aggregator = new Contract(aggregatorAbi, ddAggregator.deploy.address, provider)

    // Defines the offchain config
    const onchain_config = new Array<number>()
    const offchain_config = new Uint8Array([1])
    const offchain_config_encoded = bytesToFelts(offchain_config)
    const offchain_config_version = 2
    const config = {
      oracles: oracles.map((oracle) => {
        return {
          signer: ec.starkCurve.getStarkKey(oracle.signer),
          transmitter: oracle.transmitter.address,
        }
      }),
      f,
      onchain_config,
      offchain_config_version,
      offchain_config: offchain_config_encoded,
    }
    console.log('Encoded offchain_config: %O', offchain_config_encoded)

    // Sets the billing config
    await owner.execute(
      aggregator.populate('set_billing', {
        config: {
          observation_payment_gjuels: 1,
          transmission_payment_gjuels: 1,
          gas_base: 1,
          gas_per_signature: 1,
        },
      }),
    )

    // Sets the OCR config
    await owner.execute(aggregator.populate('set_config', config))
    console.log('Config: %O', config)

    // Gets the config details as bigints:
    //
    //   result["0"] = config_count
    //   result["1"] = block_number
    //   result["2"] = config_digest
    //
    const result = await aggregator.latest_config_details()
    const blockNumber = Number(result['1']) // we receive a bigint, but getBlock() assumes bigint = block hash, number = block number
    const configDigest = result['2']
    console.log(`Config digest: ${configDigest.toString(16)}`)
    console.log(`Block number: ${blockNumber.toString(16)}`)
    config_digest = configDigest

    // Immitate the fetch done by relay to confirm latest_config_details_works
    const block = await provider.getBlock(blockNumber)
    const txHash = block.transactions.at(0)
    if (txHash == null) {
      assert.fail('unexpectedly found no transacitons')
    }

    // Gets the transaction receipt
    const receipt = await provider.waitForTransaction(txHash)

    // Checks that the receipt has events to decode
    const events = receipt.events
    const event = events.at(0)
    if (event == null) {
      assert.fail('unexpectedly received no events')
    } else {
      console.log("Log raw 'ConfigSet' event: %O", event)
    }

    // Decodes the events
    const decodedEvents = aggregator.parseEvents(receipt)
    const decodedEvent = decodedEvents.at(0)
    if (decodedEvent == null) {
      assert.fail('unexpectedly received no decoded events')
    } else {
      console.log("Log decoded 'ConfigSet' event: %O", decodedEvent)
    }

    // Double checks that the ConfigSet event exists in the decoded event payload
    assert.isTrue(Object.prototype.hasOwnProperty.call(decodedEvent, 'ConfigSet'))
  })

  describe('OCR aggregator behavior', function () {
    const transmit = async (epochAndRound: number, answer: num.BigNumberish) => {
      // Defines helper variables
      const observations = new Array<num.BigNumberish>()
      const observersBuf = Buffer.alloc(31)
      const observationTimestamp = 1
      const juelsPerFeeCoin = 1
      const extraHash = 1
      const gasPrice = 1

      // Updates the observer state
      for (let i = 0; i < oracles.length; i++) {
        observersBuf[i] = i
        observations.push(answer)
      }

      // Converts observersBuf to a single value that will be decoded by toBN
      const observers = `0x${observersBuf.toString('hex')}`
      assert.equal(observers, OBSERVERS_HEX)

      // Defines report data
      const reportData = [
        // report_context
        config_digest,
        epochAndRound,
        extraHash,
        // raw_report
        observationTimestamp,
        observers,
        observations.length,
        ...observations,
        juelsPerFeeCoin,
        gasPrice,
      ]

      // Hashes the report data
      const reportDigest = hash.computeHashOnElements(reportData)
      console.log('Report data: %O', reportData)
      console.log(`Report digest: ${reportDigest}`)

      // Generates report signatures
      console.log('Report signatures - START')
      const signatures = []
      for (const { signer } of oracles.slice(0, f + 1)) {
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

      // Gets the first transmitter
      const transmitter = oracles.at(0)?.transmitter
      if (transmitter == null) {
        assert.fail('no oracles exist')
      }

      // Executes the transmit function on the aggregator contract
      return await transmitter.execute(
        aggregator.populate('transmit', {
          report_context: {
            config_digest,
            epoch_and_round: epochAndRound,
            extra_hash: extraHash,
          },
          observation_timestamp: observationTimestamp,
          observers,
          observations,
          juels_per_fee_coin: juelsPerFeeCoin,
          gas_price: gasPrice,
          signatures,
        }),
      )
    }

    it("should emit 'NewTransmission' event on transmit", async () => {
      // Calls the transmit function
      const { transaction_hash } = await transmit(1, 99)
      const receipt = await provider.getTransactionReceipt(transaction_hash)

      // Double checks that some events were emitted
      assert.isNotEmpty(receipt.events)
      console.log("Log raw 'NewTransmission' event: %O", receipt.events[0])

      // Decodes the events
      const decodedEvents = aggregator.parseEvents(receipt)
      const decodedEvent = decodedEvents.at(0)
      if (decodedEvent == null) {
        assert.fail('unexpectedly received no decoded events')
      } else {
        console.log("Log decoded 'NewTransmission' event: %O", decodedEvent)
      }

      // Validates the decoded event
      const e = decodedEvent['NewTransmission']
      assert.isTrue(Object.prototype.hasOwnProperty.call(decodedEvent, 'NewTransmission'))
      assert.equal(e.round_id, 1n)
      assert.equal(e.observation_timestamp, 1n)
      assert.equal(e.epoch_and_round, 1n)
      // assert.equal(e.data.reimbursement, 0n)

      // NOTICE: Leading zeros are trimmed for an encoded felt (number).
      //   To decode, the raw felt needs to be start padded up to max felt size (252 bits or < 32 bytes).
      const hexPadStart = (
        data: BigNumberish | Uint256 | ParsedStruct | BigNumberish[],
        len: number,
      ) => {
        return `0x${data.toString(16).padStart(len, '0')}`
      }

      // Validates the transmitter
      const transmitterAddr = oracles[0].transmitter.address
      const len = 32 * 2 // 32 bytes (hex)
      expect(hexPadStart(e.transmitter, len)).to.hexEqual(transmitterAddr)

      // Validates the observers and observations
      const lenObservers = OBSERVERS_MAX * 2 // 31 bytes (hex)
      assert.equal(hexPadStart(e.observers, lenObservers), OBSERVERS_HEX)
      if (Array.isArray(e.observations)) {
        assert.equal(e.observations.length, 4)
      } else {
        assert.fail(
          `property 'observations' on NewTransmission event is not an array: ${JSON.stringify(
            e,
            null,
            2,
          )}`,
        )
      }

      // Validates the config digest
      assert.equal(hexPadStart(e.config_digest, len), config_digest)
    })

    it('should transmit correctly', async () => {
      await transmit(2, 99)

      // Gets the latest round details as a map from string to bigint:
      //
      //   result["round_id"] = 2n
      //   result["answer"] = 99n
      //   result["block_num"] = 8n
      //   result["started_at"] = 1n
      //   result["updated_at"] = 1710802726n
      //
      const round = await aggregator.latest_round_data()
      assert.equal(round['round_id'], 2n)
      assert.equal(round['answer'], 99n)

      // await transmit(3, -10) // TODO: toFelt() to correctly wrap negative ints
      // ;({ round } = await aggregator.call('latest_round_data'))
      // assert.equal(round.round_id, 3)
      // assert.equal(round.answer, -10)

      try {
        await transmit(4, 1)
        expect.fail()
      } catch (err) {
        // Round should be unchanged
        const newRound = await aggregator.latest_round_data()
        assert.deepEqual(round, newRound)
      }
    })

    it('should transmit with max u128 value correctly', async () => {
      try {
        await transmit(4, UINT128_MAX)
        assert.fail('expected an error')
      } catch (err) {
        if (err instanceof LibraryError) {
          expect(err.message).to.contain('median is out of min-max range')
        } else {
          assert.fail('expected a starknet LibraryError')
        }
      }
    })

    it('payments and withdrawals', async () => {
      // set up payees
      await owner.execute(
        aggregator.populate('set_payees', {
          payees: oracles.map((oracle) => ({
            transmitter: oracle.transmitter.address,
            payee: oracle.transmitter.address, // reusing transmitter acocunts as payees for simplicity
          })),
        }),
      )

      // Several rounds happened so we are owed payment
      //
      // The aggregator.owed_payment call returns a bigint
      //
      const payee = oracles[0].transmitter
      const owed1 = await aggregator.owed_payment(payee.address)
      assert.ok(owed1 > 0n)

      const availableToValue = ([is_negative, abs_difference]: [boolean, bigint]): bigint => {
        return is_negative ? -abs_difference : abs_difference
      }

      // no funds on contract, so no LINK available for payment
      //
      // The aggregator.link_available_for_payment call returns a map:
      //
      //   result['0'] = is negative (e.g. true)
      //   result['1'] = absolute difference (e.g. 10000000006n)
      //
      let result = await aggregator.link_available_for_payment()
      assert.ok(availableToValue([result['0'], result['1']]) < 0) // should be negative: we owe payments

      // deposit LINK to contract
      await owner.execute(
        token.populate('transfer', {
          recipient: aggregator.address,
          amount: cairo.uint256(100_000_000_000n),
        }),
      )

      // we have enough funds available now
      result = await aggregator.link_available_for_payment()
      assert.ok(availableToValue([result['0'], result['1']]) > 0)

      // attempt to withdraw the payment
      await payee.execute(
        aggregator.populate('withdraw_payment', {
          transmitter: payee.address,
        }),
      )

      // balance as transferred to payee
      const balance = await token.balance_of(payee.address)
      assert.ok(owed1 === balance)

      // owed payment is now zero
      const owed2 = await aggregator.owed_payment(payee.address)
      assert.ok(owed2 === 0n)
    })
  })
})
