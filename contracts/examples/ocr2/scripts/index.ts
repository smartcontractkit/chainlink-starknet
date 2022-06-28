import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/../starknet-artifacts/contracts/${name}.cairo/${name}.json`).toString('ascii'),
  )
}

export const loadAccount = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../starknet-artifacts/account-contract-artifacts/OpenZeppelinAccount/0.1.0/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}
