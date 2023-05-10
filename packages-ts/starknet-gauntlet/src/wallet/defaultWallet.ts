import { Signer } from 'starknet'
import { IStarknetWallet } from './'
import { Env } from '../dependencies'

export class Wallet implements IStarknetWallet {
  signer: Signer
  account: string

  private constructor(keypair: string, account?: string) {
    this.signer = new Signer(keypair)
    this.account = account
  }

  static create = (pKey: string, account?: string) => {
    return new Wallet(pKey, account)
  }

  getPublicKey = async () => await this.signer.getPubKey()
  getAccountAddress = () => this.account
}

export const makeWallet = async (env: Env): Promise<IStarknetWallet> => {
  return Wallet.create(env.pk, env.account)
}
