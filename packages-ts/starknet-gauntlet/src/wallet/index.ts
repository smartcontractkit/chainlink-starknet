import { SignerInterface } from 'starknet'
import { makeWallet as makeDefaultWallet } from './defaultWallet'
import { makeWallet as makeLedgerWallet } from './ledgerWallet'
import { Env } from '../dependencies'

export interface IWallet<W> {
  wallet: W
  getPublicKey: () => Promise<string>
}

export interface IStarknetWallet extends IWallet<SignerInterface> {
  getAccountAddress: () => string
}

export const makeWallet = async (env: Env) => {
  if (env.withLedger) {
    return makeLedgerWallet(env)
  }

  return makeDefaultWallet(env)
}
