import fs from 'fs'
import { CompiledContract, json, hash } from 'starknet'
import BN from 'bn.js'

// contract hash
// required in calculation of deployment address: https://docs.starknet.io/docs/Contracts/contract-address/
// can be calculated using this formula: https://docs.starknet.io/docs/Contracts/contract-hash/
// easier to deploy an instance then get the hash from: https://alpha4.starknet.io/feeder_gateway/get_class_hash_at?contractAddress=<contract-address>
export const CONTRACT_HASH = '0x421374889bea7fd1bd78f976f6ae49d06cba62f5f879d0764fd8403c440400b'

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

export const calculateAddress = (salt: number, publicKey: string): string => {
  return hash.calculateContractAddressFromHash(salt, CONTRACT_HASH, [publicKey], 0)
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
