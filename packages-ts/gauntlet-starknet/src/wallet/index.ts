import { ec, KeyPair, Signer } from 'starknet'
export interface IWallet<W> {
  wallet: W
  sign: (message: any) => any
  getPublicKey: () => Promise<string>
}

export interface IStarknetWallet extends IWallet<Signer> {}

export const makeWallet = (rawPk?: string) => {
  return Wallet.create(rawPk)
}

class Wallet implements IStarknetWallet {
  wallet: Signer

  private constructor(keypair: KeyPair) {
    this.wallet = new Signer(keypair)
  }

  static create = (pKey: string) => {
    const keyPair = ec.getKeyPair(pKey)
    return new Wallet(keyPair)
  }

  sign = () => {}

  getPublicKey = async () => await this.wallet.getPubKey()
}
