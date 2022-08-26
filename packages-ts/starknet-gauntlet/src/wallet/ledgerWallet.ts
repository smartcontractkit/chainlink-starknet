import { IStarknetWallet } from './'
import { SignerInterface, encode } from 'starknet'
import { Abi, Invocation, InvocationsSignerDetails } from 'starknet/types'
import { TypedData, getMessageHash } from 'starknet/utils/typedData'
import { fromCallsToExecuteCalldataWithNonce } from 'starknet/utils/transaction'
import { calculcateTransactionHash, getSelectorFromName } from 'starknet/utils/hash'
import Stark, { LedgerError } from '@ledgerhq/hw-app-starknet'
import TransportNodeHid from '@ledgerhq/hw-transport-node-hid'

// EIP-2645 path
//  2645 - EIP number
//  579218131 - layer - 31 lowest bits of sha256("starkex")
//  894929996 - application - 31 lowest bits of sha256("chainlink")
const DEFAULT_PATH = "m/2645'/579218131'/894929996'/0'"
const PATH_REGEX = /^\s*m\s*\/\s*2645\s*\'\s*\/\s*579218131\s*\'\s*\/\s*(\d+)\s*\'\s*\/\s*(\d+)\s*\'$/

class LedgerSigner implements SignerInterface {
  ledgerConnector: Stark
  path: string
  publicKey: string

  private constructor(lc: Stark, path: string, publicKey: string) {
    this.ledgerConnector = lc
    this.path = path
    this.publicKey = publicKey
  }

  static create = async (path?: string) => {
    if (!path) {
      path = DEFAULT_PATH
    } else {
      const match = PATH_REGEX.exec(path)
      if (!match) {
        throw new Error(
          "Provided ledger path doesn't correspond the expected format m/2645'/579218131'/<application>'/<index>'",
        )
      }
    }

    const transport = await TransportNodeHid.create()
    const ledgerConnector = new Stark(transport)

    const response = await ledgerConnector.getPubKey(path)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to get public key: ${response.errorMessage}. Is Ledger app active?`)
    }

    const publicKey = encode.addHexPrefix(encode.buf2hex(response.publicKey).slice(2, 2 + 64))
    const signer = new LedgerSigner(ledgerConnector, path, publicKey)

    return signer
  }

  async getPubKey() {
    return this.publicKey
  }

  async signTransaction(
    transactions: Invocation[],
    transactionsDetail: InvocationsSignerDetails,
    abis?: Abi[],
  ) {
    if (abis && abis.length !== transactions.length) {
      throw new Error('ABI must be provided for each transaction or no transaction')
    }

    const calldata = fromCallsToExecuteCalldataWithNonce(transactions, transactionsDetail.nonce)

    const msgHash = calculcateTransactionHash(
      transactionsDetail.walletAddress,
      transactionsDetail.version,
      getSelectorFromName('__execute__'),
      calldata,
      transactionsDetail.maxFee,
      transactionsDetail.chainId,
    )

    return this.sign(msgHash)
  }

  async signMessage(typedData: TypedData, accountAddress: string) {
    const msgHash = getMessageHash(typedData, accountAddress)
    return this.sign(msgHash)
  }

  async sign(hash: string) {
    const response = await this.ledgerConnector.signFelt(this.path, hash, true)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to sign the message: ${response.signMessage}`)
    }

    return [
      encode.addHexPrefix(encode.buf2hex(response.r)),
      encode.addHexPrefix(encode.buf2hex(response.s)),
    ]
  }
}

export class Wallet implements IStarknetWallet {
  wallet: LedgerSigner
  account: string

  private constructor(signer: LedgerSigner, account?: string) {
    this.wallet = signer
    this.account = account
  }

  static create = async (ledgerPath?: string, account?: string) => {
    const ledgerSigner: LedgerSigner = await LedgerSigner.create(ledgerPath)
    return new Wallet(ledgerSigner, account)
  }

  getPublicKey = async () => await this.wallet.getPubKey()
  getAccountPublicKey = () => this.account
}
