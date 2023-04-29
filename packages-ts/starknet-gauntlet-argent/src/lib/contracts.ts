import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'argent_account',
}

export const loadContract = (name: CONTRACT_LIST) => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.sierra.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.casm.json`,
        'utf-8',
      ),
    ),
  }
}

export const accountContractLoader = () => loadContract(CONTRACT_LIST.ACCOUNT)
