import { TransactionResponse } from '../transaction'
import {
  SequencerProvider as StarknetProvider,
  DeployContractResponse,
  Sequencer,
  CompiledContract,
  Account,
  Call,
} from 'starknet'
import { IStarknetWallet } from '../wallet'

// TODO: Move to gauntlet-core
interface IProvider<P> {
  provider: P
  send: () => Promise<TransactionResponse>
  deployContract: (
    contract: CompiledContract,
    compiledClassHash: string,
    input: any,
    wait?: boolean,
    salt?: number,
  ) => Promise<TransactionResponse>
  signAndSend: (calls: Call[], wait?: boolean) => Promise<TransactionResponse>
}

export interface IStarknetProvider extends IProvider<StarknetProvider> {}
export const makeProvider = (
  url: string,
  wallet?: IStarknetWallet,
): IProvider<StarknetProvider> => {
  return new Provider(url, wallet)
}

export const wrapResponse = (
  provider: IStarknetProvider,
  response: Sequencer.AddTransactionResponse | DeployContractResponse,
  address?: string,
): TransactionResponse => {
  const txResponse: TransactionResponse = {
    hash: response.transaction_hash,
    // HACK: Work around the response being either AddTransactionResponse or DeployContractResponse
    address: address || (response as any).address || (response as any).contract_address,
    wait: async () => {
      // Success if does not throw
      let success: boolean
      try {
        await provider.provider.waitForTransaction(response.transaction_hash)
        txResponse.status = 'ACCEPTED'
        success = true
      } catch (e) {
        txResponse.status = 'REJECTED'
        txResponse.errorMessage = e.message
        success = false
      }
      const status = await provider.provider.getTransactionStatus(response.transaction_hash)
      txResponse.tx.code = status.tx_status as any // For some reason, starknet does not consider any other status than "TRANSACTION_RECEIVED"
      return { success }
    },
    status: 'PENDING',
    tx: response,
  }
  return txResponse
}

class Provider implements IStarknetProvider {
  provider: StarknetProvider
  account: Account

  constructor(baseUrl: string, wallet?: IStarknetWallet) {
    this.provider = new StarknetProvider({ baseUrl })
    if (wallet) {
      this.account = new Account(this.provider, wallet.getAccountAddress(), wallet.signer)
    }
  }

  setAccount(wallet: IStarknetWallet) {
    this.account = new Account(this.provider, wallet.getAccountAddress(), wallet.signer)
  }

  send = async () => {
    // Use provider to send tx and wrap it in our type
    return {} as TransactionResponse
  }

  deployContract = async (
    contract: CompiledContract,
    compiledClassHash?: string,
    input: any = [],
    wait = true,
    salt = undefined,
  ) => {
    const tx = await this.account.declareAndDeploy({
      contract,
      compiledClassHash,
      salt: salt ? '0x' + salt.toString(16) : salt, // convert number to hex or leave undefined
      // unique: false,
      ...(!!input && input.length > 0 && { constructorCalldata: input }),
    })

    const response = wrapResponse(this, tx.deploy)

    if (!wait) return response
    await response.wait()
    return response
  }

  signAndSend = async (calls: Call[], wait = false) => {
    const tx = await this.account.execute(calls)
    const response = wrapResponse(this, tx)
    if (!wait) return response

    await response.wait()
    return response
  }
}
