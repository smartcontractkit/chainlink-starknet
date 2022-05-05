export interface IWallet {
  sign: (message: any) => any
}

export const makeWallet = (rawPk?: string) => {
  return new Wallet()
}

class Wallet implements IWallet {
  constructor() {}

  sign = () => {}
}
