import { expect } from "chai";
import { starknet } from "hardhat";
import { ec, encode, hash, number, stark, KeyPair } from "starknet";
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
      initial_supply: { high: 0n, low: 1000n },
      recipient: BigInt(owner.starknetContract.address),
      owner: BigInt(owner.starknetContract.address),
    })
    
    aggregator = await aggregatorContractFactory.deploy({
      owner: BigInt(owner.starknetContract.address),
      link: BigInt(token.address),
      min_answer: minAnswer,
      max_answer: maxAnswer,
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
    answer: number
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

  it("sets up the environment", async() => {
      await transmit(1, 99)
      await transmit(2, -10)
      await transmit(3, -100)
  });

})