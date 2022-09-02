import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(`${__dirname}/../starknet-artifacts/contracts/${name}.cairo/${name}.json`)
      .toString('ascii'),
  )
}

export const loadContract_Account = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../node_modules/@shardlabs/starknet-hardhat-plugin/dist/account-contract-artifacts/OpenZeppelinAccount/0.2.1/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}
