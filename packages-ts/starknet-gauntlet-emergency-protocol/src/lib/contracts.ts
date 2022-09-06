import fs from 'fs'
import { CompiledContract, json } from 'starknet'
import { ContractFactory } from 'ethers'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'sequencer_uptime_feed',
  STARKNET_VALIDATOR = 'StarknetValidator',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/../../contract_artifacts/abi/${name}.json`).toString('ascii'),
  )
}

export const loadStarknetValidatorContract = (name: CONTRACT_LIST): ContractFactory => {
  const abi = JSON.parse(
    fs.readFileSync(`${__dirname}/../../contract_artifacts/abi/${name}.json`).toString('ascii'),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}

export const uptimeFeedContractLoader = () => loadContract(CONTRACT_LIST.SEQUENCER_UPTIME_FEED)
export const starknetValidatorContractLoader = () =>
  loadStarknetValidatorContract(CONTRACT_LIST.STARKNET_VALIDATOR)
