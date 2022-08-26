import { SignerInterface } from 'starknet'
import { Wallet as DefaultWallet } from './defaultWallet'
import { Wallet as LedgerWallet } from './ledgerWallet'

export interface IWallet<W> {
  wallet: W
  getPublicKey: () => Promise<string>
}

export interface IStarknetWallet extends IWallet<SignerInterface> {
  getAccountPublicKey: () => string
}

export const makeWallet = async (
  withLedger: boolean,
  ledgerPath?: string,
  rawPk?: string,
  account?: string,
): Promise<IStarknetWallet> => {
  if (withLedger) {
    return await LedgerWallet.create(ledgerPath, account)
  }

  return DefaultWallet.create(rawPk, account)
}
