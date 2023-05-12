import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'argent_account',
}

export const loadContract = (name: CONTRACT_LIST): any => {
  return {
    contract: json.parse(
      fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('utf8'),
    ),
  }
}

export const accountContractLoader = () => loadContract(CONTRACT_LIST.ACCOUNT)
