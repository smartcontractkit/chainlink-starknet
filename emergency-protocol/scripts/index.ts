import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/../starknet-artifacts/contracts/cairo/${name}.cairo/${name}.json`).toString('ascii'),
  )
}

export const loadAccount = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../node_modules/@shardlabs/starknet-hardhat-plugin/dist/account-contract-artifacts/OpenZeppelinAccount/0.2.0/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}
