import fs from 'fs'
import { CompiledContract, json, hash } from 'starknet'
import BN from 'bn.js'

export enum CONTRACT_LIST {
  ACCOUNT = 'Account',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-open-zeppelin/artifacts/0.5.0/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const accountContractLoader = () => loadContract(CONTRACT_LIST.ACCOUNT)

// use bignumber libraries to assert addresses are equal
// handles prepending 0s
export const equalAddress = (addr0: string, addr1: string): boolean => {
  let a0 = new BN(removePrefix(addr0), 16)
  let a1 = new BN(removePrefix(addr1), 16)

  return a0.cmp(a1) == 0
}

const removePrefix = (addr: string): string => {
  return addr.replace(/^(0x)/, '')
}
