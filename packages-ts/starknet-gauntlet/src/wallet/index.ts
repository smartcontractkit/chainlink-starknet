import { SignerInterface } from 'starknet'
import { Wallet as DefaultWallet } from './defaultWallet'
import { Wallet as LedgerWallet } from './ledgerWallet'

export interface IWallet<W> {
  wallet: W
  sign: (message: any) => any
  getPublicKey: () => Promise<string>
}

export interface IStarknetWallet extends IWallet<SignerInterface> {
  getAccountPublicKey: () => string
}

export const makeWallet = async (withLedger: boolean, rawPk?: string, account?: string): Promise<IStarknetWallet> => {
  if (withLedger) {
    return await LedgerWallet.create()
  }

  return DefaultWallet.create(rawPk, account)
}
