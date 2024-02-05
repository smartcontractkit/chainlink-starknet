import fs from 'fs'
import { ContractFactory } from 'ethers'
import { loadContract } from '@chainlink/starknet-gauntlet'
import * as accessControllerArtifact from '@chainlink/evm-gauntlet-common/artifacts/evm/SimpleWriteAccessController.json'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'SequencerUptimeFeed',
  STARKNET_VALIDATOR = 'starknet_validator',
  L1_ACCESS_CONTROLLER = 'AccessController',
  L1_MOCK_AGGREGATOR = 'MockAggregator',
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

export const L1AccessControllerContractLoader = () => {
  return new ContractFactory(
    accessControllerArtifact.compilerOutput?.abi,
    accessControllerArtifact.compilerOutput?.evm?.bytecode,
  )
}

export const L1MockAggregatorLoader = () => {
  const abi = JSON.parse(
    fs.readFileSync(
      `${__dirname}/../../../../contracts/artifacts/solidity/mocks/MockAggregator.sol/MockAggregator.json`,
      'utf-8',
    ),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}
