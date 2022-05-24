import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  TOKEN = 'oz_erc20',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('ascii'))
}

export const contractLoader = () => loadContract(CONTRACT_LIST.TOKEN)
