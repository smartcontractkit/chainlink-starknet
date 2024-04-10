import { IStarknetWallet, Env } from '@chainlink/starknet-gauntlet'
import { encode, Signature, ec, Signer } from 'starknet'
import { Stark as LedgerClient, LedgerError } from '@ledgerhq/hw-app-starknet'

// EIP-2645 path
//  2645 - EIP number
//  579218131 - layer - 31 lowest bits of sha256("starkex")
//  894929996 - application - 31 lowest bits of sha256("chainlink")
export const DEFAULT_LEDGER_PATH = "m/2645'/579218131'/894929996'/0'"
export const LEDGER_PATH_REGEX = /^\s*m\s*\/\s*2645\s*\'\s*\/\s*579218131\s*\'\s*\/\s*(\d+)\s*\'\s*\/\s*(\d+)\s*\'$/

class LedgerSigner extends Signer {
  client: LedgerClient
  path: string
  publicKey: string

  private constructor(client: LedgerClient, path: string) {
    super()
    this.client = client
    this.path = path
  }

  static create = async (path?: string): Promise<LedgerSigner> => {
    if (!path) {
      path = DEFAULT_LEDGER_PATH
    } else {
      const match = LEDGER_PATH_REGEX.exec(path)
      if (!match) {
        throw new Error(
          "Provided ledger path doesn't correspond the expected format m/2645'/579218131'/<application>'/<index>'",
        )
      }
    }

    // work around jest reimporting the package
    // package is only allowed to be imported once
    const Transport = (await import('@ledgerhq/hw-transport-node-hid')).default
    const transport = await Transport.create()
    const ledgerConnector = new LedgerClient(transport)

    const signer = new LedgerSigner(ledgerConnector, path)

    return signer
  }

  public async getPubKey(): Promise<string> {
    // memoize to avoid redundant calls
    if (this.publicKey) return this.publicKey

    const response = await this.client.getPubKey(this.path)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to get public key: ${response.errorMessage}. Is Ledger app active?`)
    }

    this.publicKey = encode.addHexPrefix(encode.buf2hex(response.publicKey).slice(2, 2 + 64))

    return this.publicKey
  }

  async signRaw(hash: string): Promise<Signature> {
    const response = await this.client.sign(this.path, hash, true)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to sign the message: ${response.errorMessage}`)
    }

    // TODO: console log the hash so user can verify on ledger

    return new ec.starkCurve.Signature(
      BigInt(encode.addHexPrefix(encode.buf2hex(response.r))),
      BigInt(encode.addHexPrefix(encode.buf2hex(response.s))),
    )
  }
}

export class Wallet implements IStarknetWallet {
  signer: LedgerSigner
  account: string

  private constructor(signer: LedgerSigner, account?: string) {
    this.signer = signer
    this.account = account
  }

  static create = async (ledgerPath?: string, account?: string): Promise<Wallet> => {
    const ledgerSigner: LedgerSigner = await LedgerSigner.create(ledgerPath)
    return new Wallet(ledgerSigner, account)
  }

  getPublicKey = async (): Promise<string> => await this.signer.getPubKey()
  getAccountAddress = (): string => this.account
}

export const makeWallet = async (env: Env): Promise<IStarknetWallet> => {
  return Wallet.create(env.ledgerPath, env.account)
}
