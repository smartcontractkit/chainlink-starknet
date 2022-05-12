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

  // A wallet is a contract. If the pkey is not provided, we cannot generate a random one withouth deploying the acc contract
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
