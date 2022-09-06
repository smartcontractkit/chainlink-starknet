import fs from 'fs'
import { CompiledContract, json } from 'starknet'
import { ContractFactory } from 'ethers'

// todo: fix name mapping
// l1_token_bridge = token_bridge
// l2_token_bridge = StarknetERC20Bridge
export enum CONTRACT_LIST {
  TOKEN = 'ERC20',
  L1_BRIDGE = 'l1_token_bridge',
  L2_BRIDGE = 'l2_token_bridge',
}

export const loadL2Contract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-contracts/artifacts/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadL1BridgeContract = (name: CONTRACT_LIST): ContractFactory => {
  const abi = JSON.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/internals-starkgate-contracts/artifacts/0.0.3/eth/StarknetERC20Bridge.json`,
      )
      .toString('ascii'),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}

export const tokenContractLoader = () => loadL2Contract(CONTRACT_LIST.TOKEN)
export const l1BridgeContractLoader = () => loadL1BridgeContract(CONTRACT_LIST.L1_BRIDGE)
export const l2BridgeContractLoader = () => loadL2Contract(CONTRACT_LIST.L2_BRIDGE)
