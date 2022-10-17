import { ec, KeyPair, Signer } from 'starknet'
import { IStarknetWallet } from './'
import { Env } from '../dependencies'

export class Wallet implements IStarknetWallet {
  signer: Signer
  account: string

  private constructor(keypair: KeyPair, account?: string) {
    this.signer = new Signer(keypair)
    this.account = account
  }

  static create = (pKey: string, account?: string) => {
    const keyPair = ec.getKeyPair(pKey)
    return new Wallet(keyPair, account)
  }

  getPublicKey = async () => await this.signer.getPubKey()
  getAccountAddress = () => this.account
}

export const makeWallet = async (env: Env): Promise<IStarknetWallet> => {
  return Wallet.create(env.pk, env.account)
}
