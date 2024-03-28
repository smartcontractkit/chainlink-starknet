import fs from 'fs'
import { ContractFactory } from 'ethers'
import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'SequencerUptimeFeed',
  STARKNET_VALIDATOR = 'starknet_validator',
}

export const uptimeFeedContractLoader = () => {
  return loadContract(CONTRACT_LIST.SEQUENCER_UPTIME_FEED)
}

export const starknetValidatorContractLoader = (): ContractFactory => {
  const abi = JSON.parse(
    fs.readFileSync(
      `${__dirname}/../../../../contracts/artifacts/solidity/emergency/StarknetValidator.sol/StarknetValidator.json`,
      'utf-8',
    ),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}
