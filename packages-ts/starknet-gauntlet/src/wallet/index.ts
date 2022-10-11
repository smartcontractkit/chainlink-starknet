import { SignerInterface } from 'starknet'
import { makeWallet } from './defaultWallet'

export interface IWallet<W> {
  wallet: W
  getPublicKey: () => Promise<string>
}

export interface IStarknetWallet extends IWallet<SignerInterface> {
  getAccountAddress: () => string
}

export { makeWallet }
