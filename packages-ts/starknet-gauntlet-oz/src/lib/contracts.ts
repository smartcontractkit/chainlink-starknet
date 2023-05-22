import fs from 'fs'
import { json } from 'starknet'
import BN from 'bn.js'

export enum CONTRACT_LIST {
  ACCOUNT = 'Account',
}

export const accountContractLoader = () => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starkgate-open-zeppelin/artifacts/0.5.0/Account.cairo/Account.json`,
        'utf8',
      ),
    ),
  }
}

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
