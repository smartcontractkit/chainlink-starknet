import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  TOKEN = 'ERC20',
  BRIDGE = 'token_bridge',
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
export const bridgeContractLoader = () => loadContract(CONTRACT_LIST.BRIDGE)
