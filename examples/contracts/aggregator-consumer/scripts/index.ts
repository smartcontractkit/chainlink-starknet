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
        `${__dirname}/../node_modules/@shardlabs/starknet-hardhat-plugin/dist/account-contract-artifacts/OpenZeppelinAccount/b27101eb826fae73f49751fa384c2a0ff3377af2/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}
