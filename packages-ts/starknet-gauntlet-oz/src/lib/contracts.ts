import fs from 'fs'
import { json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'Account',
}

export const accountContractLoader = () => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_Account.contract_class.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_Account.compiled_contract_class.json`,
        'utf-8',
      ),
    ),
  }
}
