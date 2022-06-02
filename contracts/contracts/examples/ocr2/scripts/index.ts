import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export const loadORC2ConsumerContract = (name: string): CompiledContract => {
    return json.parse(
      fs
        .readFileSync(
          `${__dirname}/../starknet-artifacts/contracts/${name}.cairo/${name}.json`,
        )
        .toString('ascii'),
    )
  }