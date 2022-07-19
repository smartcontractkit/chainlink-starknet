import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  ACCOUNT = 'Account',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-open-zeppelin/artifacts/0.2.1/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const accountContractLoader = () => loadContract(CONTRACT_LIST.ACCOUNT)
