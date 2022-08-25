import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'argent_account',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('ascii'),
  )
}

export const accountContractLoader = () => loadContract(CONTRACT_LIST.ACCOUNT)
