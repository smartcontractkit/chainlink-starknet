import { ec, KeyPair, Signer } from 'starknet'
import { IStarknetWallet } from './'

export class Wallet implements IStarknetWallet {
  wallet: Signer
  account: string

  private constructor(keypair: KeyPair, account?: string) {
    this.wallet = new Signer(keypair)
    this.account = account
  }

  static create = (pKey: string, account?: string) => {
    const keyPair = ec.getKeyPair(pKey)
    return new Wallet(keyPair, account)
  }

  getPublicKey = async () => await this.wallet.getPubKey()
  getAccountPublicKey = () => this.account
}
