import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  EXAMPLE = 'example',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../../contract_artifacts/abi/${name}.json`).toString('ascii'))
}

export const tokenContractLoader = () => loadContract(CONTRACT_LIST.EXAMPLE)
