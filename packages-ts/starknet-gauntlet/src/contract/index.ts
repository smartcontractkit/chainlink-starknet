import fs from 'fs'
import { json } from 'starknet'
import { LoadContractResult } from '../commands/base/executeCommand'

export const loadContract = (name: string): LoadContractResult => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.contract_class.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.compiled_contract_class.json`,
        'utf-8',
      ),
    ),
  }
}
