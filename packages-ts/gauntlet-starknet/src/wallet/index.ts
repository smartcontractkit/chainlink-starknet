import { ec, KeyPair, Signer, stark } from 'starknet'
export interface IWallet<W> {
  wallet: W
  sign: (message: any) => any
}

export interface IStarknetWallet extends IWallet<Signer> {}

export const makeWallet = (rawPk?: string) => {
  return Wallet.create(rawPk)
}

class Wallet implements IStarknetWallet {
  wallet: Signer

  private constructor(keypair?: KeyPair) {
    this.wallet = new Signer(keypair)
  }

  static create = (keypair?: KeyPair) => {
    if (!keypair) {
      const pk = stark.randomAddress()
      return new Wallet(ec.getKeyPair(pk))
    }
    return new Wallet(keypair)
  }

  sign = () => {}

  getPublicKey = async () => await this.wallet.getPubKey()
}
