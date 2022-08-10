import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  TOKEN = 'ERC20',
  L1_BRIDGE = 'l1_token_bridge',
  L2_BRIDGE = 'l2_token_bridge',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-contracts/artifacts/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const contractLoader = () => loadContract(CONTRACT_LIST.TOKEN)
export const l1BridgeContractLoader = () => loadContract(CONTRACT_LIST.L1_BRIDGE)
export const l2BridgeContractLoader = () => loadContract(CONTRACT_LIST.L2_BRIDGE)
