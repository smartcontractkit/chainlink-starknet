import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  MULTISIG = 'Multisig',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/starsign-multisig/starknet-artifacts/contracts/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const contractLoader = () => loadContract(CONTRACT_LIST.MULTISIG)
