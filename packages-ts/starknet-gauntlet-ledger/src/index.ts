import { IStarknetWallet, Env } from '@chainlink/starknet-gauntlet'
import { SignerInterface, encode, DeployAccountSignerDetails, DeclareSignerDetails } from 'starknet'
import { Abi, Signature, Call, InvocationsSignerDetails } from 'starknet/types'
import { TypedData, getMessageHash } from 'starknet/utils/typedData'
import { fromCallsToExecuteCalldataWithNonce } from 'starknet/utils/transaction'
import { calculateTransactionHash } from 'starknet/utils/hash'
import { Stark, LedgerError } from '@ledgerhq/hw-app-starknet'

// EIP-2645 path
//  2645 - EIP number
//  579218131 - layer - 31 lowest bits of sha256("starkex")
//  894929996 - application - 31 lowest bits of sha256("chainlink")
export const DEFAULT_LEDGER_PATH = "m/2645'/579218131'/894929996'/0'"
export const LEDGER_PATH_REGEX = /^\s*m\s*\/\s*2645\s*\'\s*\/\s*579218131\s*\'\s*\/\s*(\d+)\s*\'\s*\/\s*(\d+)\s*\'$/

// TODO: why are we copying this rather than using https://github.com/yogh333/starknetjs-signer-ledger/blob/main/src/ledger-signer.ts
class LedgerSigner implements SignerInterface {
  client: Stark
  path: string
  publicKey: string

  private constructor(client: Stark, path: string, publicKey: string) {
    this.client = client
    this.path = path
    this.publicKey = publicKey
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
    const ledgerConnector = new Stark(transport)

    const response = await ledgerConnector.getPubKey(path)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to get public key: ${response.errorMessage}. Is Ledger app active?`)
    }

    const publicKey = encode.addHexPrefix(encode.buf2hex(response.publicKey).slice(2, 2 + 64))
    const signer = new LedgerSigner(ledgerConnector, path, publicKey)

    return signer
  }

  async getPubKey(): Promise<string> {
    return this.publicKey
  }

  async signTransaction(
    transactions: Call[],
    transactionsDetail: InvocationsSignerDetails,
    abis?: Abi[],
  ): Promise<Signature> {
    const calldata = fromCallsToExecuteCalldataWithNonce(transactions, transactionsDetail.nonce)

    const msgHash = calculateTransactionHash(
      transactionsDetail.walletAddress,
      transactionsDetail.version,
      calldata,
      transactionsDetail.maxFee,
      transactionsDetail.chainId,
      transactionsDetail.nonce,
    )

    return this.sign(msgHash)
  }

  async signDeployAccountTransaction(transaction: DeployAccountSignerDetails): Promise<Signature> {
    // TODO: implement this
    return []
  }

  async signDeclareTransaction(transaction: DeclareSignerDetails): Promise<Signature> {
    // TODO: implement this
    return []
  }

  async signMessage(typedData: TypedData, accountAddress: string): Promise<Signature> {
    const msgHash = getMessageHash(typedData, accountAddress)
    return this.sign(msgHash)
  }

  async sign(hash: string): Promise<Signature> {
    const response = await this.client.sign(this.path, hash, true)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to sign the message: ${response.errorMessage}`)
    }

    return [
      encode.addHexPrefix(encode.buf2hex(response.r)),
      encode.addHexPrefix(encode.buf2hex(response.s)),
    ]
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
