import { assert, expect } from "chai";
import { starknet } from "hardhat";
import { constants, ec, encode, hash, number, uint256, stark, KeyPair } from "starknet";
import { BigNumberish } from "starknet/utils/number";
import {
  Account,
  StarknetContract,
  StarknetContractFactory,
} from "hardhat/types/runtime";
import { TIMEOUT } from "./constants";

interface Oracle {
  signer: KeyPair,
  transmitter: Account,
}

// Required to convert negative values into [0, PRIME) range
function toFelt(int: number | BigNumberish): BigNumberish {
  let prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME));
  return number.toBN(int).umod(prime)
}

describe("aggregator.cairo", function () {
  this.timeout(TIMEOUT);
  
  let aggregatorContractFactory: StarknetContractFactory;
  // let accountContractFactory: StarknetContractFactory;
  let tokenContractFactory: StarknetContractFactory;
  
  let owner: Account;
  let token: StarknetContract;
  let aggregator: StarknetContract;
  
  let minAnswer = -10
  let maxAnswer = 1000000000

  let f = 1;
  let n = 3 * f + 1;
  let oracles: Oracle[] = []
  let config_digest: number;

  before(async function() {
    // assumes contract.cairo and events.cairo has been compiled
    aggregatorContractFactory = await starknet.getContractFactory("aggregator");
    tokenContractFactory = await starknet.getContractFactory("token");

    // can also be declared as
    // account = (await starknet.deployAccount("OpenZeppelin")) as OpenZeppelinAccount
    // if imported from hardhat/types/runtime"
    owner = await starknet.deployAccount("OpenZeppelin");
    
    token = await tokenContractFactory.deploy({
      name: starknet.shortStringToBigInt("LINK Token"),
      symbol: starknet.shortStringToBigInt("LINK"),
      decimals: 18,
      initial_supply: { high: 1n, low: 0n },
      recipient: BigInt(owner.starknetContract.address),
      owner: BigInt(owner.starknetContract.address),
    })
    
    aggregator = await aggregatorContractFactory.deploy({
      owner: BigInt(owner.starknetContract.address),
      link: BigInt(token.address),
      min_answer: toFelt(minAnswer),
      max_answer: toFelt(maxAnswer),
      billing_access_controller: 0, // TODO: billing AC
      decimals: 8,
      description: starknet.shortStringToBigInt("FOO/BAR")
    })
    
    let futures = [];
    let generateOracle = async () => {
      let transmitter = await starknet.deployAccount("OpenZeppelin");
      return {
        signer: ec.genKeyPair(),
        transmitter,
        // payee
      };
    };
    for (let i = 0; i < n; i++) {
      futures.push(generateOracle());
    }
    oracles = await Promise.all(futures);
    
    let onchain_config = 1;
    let offchain_config_version = 2; // TODO: assert == 2 in contract
    let offchain_config = [1];
    
    await owner.invoke(aggregator, "set_config", {
      oracles: oracles.map((oracle) => {
        return {
          signer: number.toBN(ec.getStarkKey(oracle.signer)),
          transmitter: oracle.transmitter.starknetContract.address,
        }
      }),
      f,
      onchain_config,
      offchain_config_version,
      offchain_config,
    })
    
    let result = await aggregator.call("latest_config_details")
    config_digest = result.config_digest
   
  });
  
  let transmit = async (
    epoch_and_round: number,
    answer: BigNumberish
  ): Promise<any> => {
    
    let extra_hash = 1
    let observation_timestamp = 1
    let juels_per_fee_coin = 1

    let observers_buf = Buffer.alloc(31)
    let observations = []

    for (const [index, _oracle] of oracles.entries()) {
      observers_buf[index] = index
      observations.push(answer)
    }
    
    // convert to a single value that will be decoded by toBN
    let observers = `0x${observers_buf.toString('hex')}`

    let msg = hash.computeHashOnElements([
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
    ])
    
    let signatures = []

    for (let oracle of oracles.slice(0, f + 1)) {
      let [r, s] = ec.sign(oracle.signer, msg)
      signatures.push({ r, s, public_key: number.toBN(ec.getStarkKey(oracle.signer)) })
    }
    
    let transmitter = oracles[0].transmitter
    
    return await transmitter.invoke(aggregator, "transmit", {
      report_context: {
        config_digest,
        epoch_and_round,
        extra_hash,
      },
      observation_timestamp,
      observers,
      observations,
      juels_per_fee_coin,
      signatures,
    })
  }

  it("billing config", async() => {
    await owner.invoke(aggregator, "set_billing", {
      config: {
        observation_payment_gjuels: 1,
        transmission_payment_gjuels: 5,
      }
    })
  })


  it("transmission", async() => {
      await transmit(1, 99);
      let { round } = await aggregator.call("latest_round_data")
      assert.equal(round.round_id, 1)
      assert.equal(round.answer, 99)

      await transmit(2, toFelt(-10));
      ({ round } = await aggregator.call("latest_round_data"))
      assert.equal(round.round_id, 2)
      assert.equal(round.answer, -10)

      try {
        await transmit(3, -100)
        expect.fail()
      } catch(err: any) {
        // Round should be unchanged
        let { round: new_round } = await aggregator.call("latest_round_data")
        assert.deepEqual(round, new_round)
      }

  });

  it("payee management", async() => {
    let payees = oracles.map((oracle) => ({
      transmitter: oracle.transmitter.starknetContract.address,
      payee: oracle.transmitter.starknetContract.address, // reusing transmitter acocunts as payees for simplicity
    }))
    // call set_payees, should succeed because all payees are zero
    await owner.invoke(aggregator, "set_payees", { payees })
    // call set_payees, should succeed because values are unchanged
    await owner.invoke(aggregator, "set_payees", { payees })

    let oracle = oracles[0].transmitter;
    let transmitter = oracle.starknetContract.address
    let payee = transmitter

    let proposed_oracle = oracles[1].transmitter
    let proposed_transmitter = proposed_oracle.starknetContract.address
    let proposed_payee = proposed_transmitter

    // can't transfer to self
    try {
      await oracle.invoke(aggregator, "transfer_payeeship", {
        transmitter,
        proposed: payee,
      })
      expect.fail()
    } catch(err: any) {
      // TODO: expect(err.message).to.contain("");
    }

    // only payee can transfer
    try {
      await proposed_oracle.invoke(aggregator, "transfer_payeeship", {
        transmitter,
        proposed: proposed_payee,
      })
      expect.fail()
    } catch(err: any) {
    }

    // successful transfer
    await oracle.invoke(aggregator, "transfer_payeeship", {
      transmitter,
      proposed: proposed_payee,
    })

    // only proposed payee can accept
    try {
      await oracle.invoke(aggregator, "accept_payeeship", { transmitter })
      expect.fail()
    } catch(err: any) {
    }

    // successful accept
    await proposed_oracle.invoke(aggregator, "accept_payeeship", { transmitter })
  })

  it("payments and withdrawals", async() => {
    let oracle = oracles[0];
    // NOTE: previous test changed oracle0's payee to oracle1
    let payee = oracles[1].transmitter

    let { amount: owed } = await payee.call(aggregator, "owed_payment", {
      transmitter: oracle.transmitter.starknetContract.address
    })
    // several rounds happened so we are owed payment
    assert.ok(owed > 0)

    // no funds on contract, so no LINK available for payment
    let { available } = await aggregator.call("link_available_for_payment")
    assert.ok(available < 0) // should be negative: we owe payments

    // deposit LINK to contract
    await owner.invoke(token, "transfer", {
      recipient: aggregator.address,
      amount: uint256.bnToUint256(100_000_000_000)
    })

    // we have enough funds available now
    available = (await aggregator.call("link_available_for_payment")).available
    assert.ok(available > 0)

    // attempt to withdraw the payment
    await payee.invoke(aggregator, "withdraw_payment", {
      transmitter: oracle.transmitter.starknetContract.address
    })
    
    // balance as transferred to payee
    let { balance } = await token.call("balanceOf", {
      account: payee.starknetContract.address
    })
    
    assert.ok(number.toBN(owed).eq(uint256.uint256ToBN(balance)));

    // owed payment is now zero
    owed = (await payee.call(aggregator, "owed_payment", {
      transmitter: oracle.transmitter.starknetContract.address
    })).amount
    assert.ok(owed == 0)
    
  })

})