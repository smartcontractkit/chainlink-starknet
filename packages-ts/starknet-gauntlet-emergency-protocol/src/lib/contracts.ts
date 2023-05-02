import fs from 'fs'
import { CompiledContract, json } from 'starknet'
import { ContractFactory } from 'ethers'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'sequencer_uptime_feed',
  STARKNET_VALIDATOR = 'starknet_validator',
}

export const uptimeFeedContractLoader = (): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed/sequencer_uptime_feed.cairo/sequencer_uptime_feed.json`,
      )
      .toString('ascii'),
  )
}

export const starknetValidatorContractLoader = (): ContractFactory => {
  const abi = JSON.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/artifacts/solidity/emergency/StarknetValidator.sol/StarknetValidator.json`,
      )
      .toString('ascii'),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}
