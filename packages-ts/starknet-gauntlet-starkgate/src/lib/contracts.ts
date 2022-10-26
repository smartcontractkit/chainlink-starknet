import fs from 'fs'
import { CompiledContract, json } from 'starknet'
import { ContractFactory } from 'ethers'

export enum CONTRACT_LIST {
  TOKEN = 'token',
  L1_BRIDGE = 'l1_token_bridge',
  L2_BRIDGE = 'l2_token_bridge',
}

// todo: remove when stable contract artifacts release available
const CONTRACT_NAME_TO_ARTIFACT_NAME = {
  [CONTRACT_LIST.L1_BRIDGE]: 'StarknetERC20Bridge',
  [CONTRACT_LIST.L2_BRIDGE]: 'token_bridge',
  [CONTRACT_LIST.TOKEN]: 'ERC20',
}

export const loadL2Contract = (name: CONTRACT_LIST): CompiledContract => {
  const artifactName = CONTRACT_NAME_TO_ARTIFACT_NAME[name]
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-contracts/artifacts/${artifactName}.cairo/${artifactName}.json`,
      )
      .toString('ascii'),
  )
}

export const loadL1BridgeContract = (name: CONTRACT_LIST): ContractFactory => {
  const artifactName = CONTRACT_NAME_TO_ARTIFACT_NAME[name]
  const abi = JSON.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/internals-starkgate-contracts/artifacts/0.0.3/eth/${artifactName}.json`,
      )
      .toString('ascii'),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}

export const tokenContractLoader = () => loadL2Contract(CONTRACT_LIST.TOKEN)
export const l1BridgeContractLoader = () => loadL1BridgeContract(CONTRACT_LIST.L1_BRIDGE)
export const l2BridgeContractLoader = () => loadL2Contract(CONTRACT_LIST.L2_BRIDGE)
