import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  TOKEN = 'token_example',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('ascii'))
}

export const tokenContract = loadContract(CONTRACT_LIST.TOKEN)
