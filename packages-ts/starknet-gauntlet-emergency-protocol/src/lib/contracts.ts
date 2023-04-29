import fs from 'fs'
import { json } from 'starknet'
import { ContractFactory } from 'ethers'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'sequencer_uptime_feed',
  STARKNET_VALIDATOR = 'starknet_validator',
}

export const uptimeFeedContractLoader = () => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_SequencerUptimeFeed.sierra.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_SequencerUptimeFeed.casm.json`,
        'utf-8',
      ),
    ),
  }
}

export const starknetValidatorContractLoader = (): ContractFactory => {
  const abi = JSON.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/artifacts/solidity/emergency/StarknetValidator.sol/StarknetValidator.json`,
        'utf-8'
      )
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}
