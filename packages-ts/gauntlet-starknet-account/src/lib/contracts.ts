import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'account',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('ascii'))
}

export const accountContract = loadContract(CONTRACT_LIST.ACCOUNT)
