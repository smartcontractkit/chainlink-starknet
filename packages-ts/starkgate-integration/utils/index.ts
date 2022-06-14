import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export const loadStarkgateContract = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../node_modules/@chainlink-dev/starkgate-contracts/artifacts/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}
