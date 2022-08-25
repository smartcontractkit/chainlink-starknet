import { IStarknetWallet } from './'
import { SignerInterface, encode } from 'starknet'
import { Abi, Invocation, InvocationsSignerDetails } from 'starknet/types';
import { TypedData, getMessageHash } from 'starknet/utils/typedData';
import { fromCallsToExecuteCalldataWithNonce } from 'starknet/utils/transaction';
import { calculcateTransactionHash, getSelectorFromName } from 'starknet/utils/hash';
import Stark from "@ledgerhq/hw-app-starknet";
import { LedgerError } from "@ledgerhq/hw-app-starknet";
import TransportNodeHid from "@ledgerhq/hw-transport-node-hid";

// EIP-2645 path
const PATH = "m/2645'/579218131'/0'/0'";

function toHexString(byteArray: Uint8Array): string {
  return Array.from(byteArray, function (byte) {
    return `0${byte.toString(16)}`.slice(-2);
  }).join('');
}

class LedgerSigner implements SignerInterface {
  protected ledgerConnector: Stark

  private constructor(lc: Stark) {
    this.ledgerConnector = lc
  }

  static create = async () => {
    const transport = await TransportNodeHid.create()
    const ledgerConnector = new Stark(transport)
    return new LedgerSigner(ledgerConnector)
  }

  async getPubKey() {
    const response = await this.ledgerConnector.getPubKey(PATH)
    if (response.returnCode != LedgerError.NoErrors) {
        throw new Error(`Unable to get public key: ${response.errorMessage}. Is Ledger app active?`)
    }

    return `0x${toHexString(response.publicKey).slice(2, 2 + 64)}`
  }

  async signTransaction(transactions: Invocation[], transactionsDetail: InvocationsSignerDetails, abis?: Abi[]) {
    if (abis && abis.length !== transactions.length) {
      throw new Error('ABI must be provided for each transaction or no transaction')
    }

    const calldata = fromCallsToExecuteCalldataWithNonce(transactions, transactionsDetail.nonce);

    const msgHash = calculcateTransactionHash(
      transactionsDetail.walletAddress,
      transactionsDetail.version,
      getSelectorFromName('__execute__'),
      calldata,
      transactionsDetail.maxFee,
      transactionsDetail.chainId
    );

    return this.sign(msgHash)
  }

  async signMessage(typedData: TypedData, accountAddress: string) {
    const msgHash = getMessageHash(typedData, accountAddress);
    return this.sign(msgHash)
  }

  async sign(hash: string) {
    const response = await this.ledgerConnector.signFelt(PATH, hash, true)
    if (response.returnCode != LedgerError.NoErrors) {
      throw new Error(`Unable to sign the message: ${response.signMessage}`)
    }
    
    return [
      encode.addHexPrefix(toHexString(response.r)),
      encode.addHexPrefix(toHexString(response.s)),
    ];
  }
}

export class Wallet implements IStarknetWallet {
  wallet: LedgerSigner
  account: string

  private constructor(signer: LedgerSigner, account?: string) {
    this.wallet = signer
    this.account = account
  }

  static create = async (account?: string) => {
    const ledgerSigner: LedgerSigner = await LedgerSigner.create()
    return new Wallet(ledgerSigner, account)
  }

  sign = () => {}

  getPublicKey = async () => await this.wallet.getPubKey()
  getAccountPublicKey = () => this.account
}

