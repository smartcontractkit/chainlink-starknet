import { TransactionResponse } from '../transaction'
import {
  SequencerProvider as StarknetProvider,
  DeployContractResponse,
  api,
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
    input: any,
    wait?: boolean,
    salt?: number,
  ) => Promise<TransactionResponse>
  signAndSend: (
    accountAddress: string,
    wallet: IStarknetWallet,
    calls: Call[],
  ) => Promise<TransactionResponse>
}

export interface IStarknetProvider extends IProvider<StarknetProvider> {}

export const makeProvider = (url: string): IProvider<StarknetProvider> => {
  return new Provider(url)
}

export const wrapResponse = (
  provider: IStarknetProvider,
  response: api.Sequencer.AddTransactionResponse | DeployContractResponse,
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
        success =
          (await provider.provider.waitForTransaction(response.transaction_hash)) === undefined
        txResponse.status = 'ACCEPTED'
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

  constructor(baseUrl: string) {
    this.provider = new StarknetProvider({ baseUrl })
  }

  send = async () => {
    // Use provider to send tx and wrap it in our type
    return {} as TransactionResponse
  }

  deployContract = async (
    contract: CompiledContract,
    input: any = [],
    wait = true,
    salt = undefined,
  ) => {
    const tx = await this.provider.deployContract({
      contract,
      addressSalt: salt ? '0x' + salt.toString(16) : salt, // convert number to hex or leave undefined
      ...(!!input && input.length > 0 && { constructorCalldata: input }),
    })

    const response = wrapResponse(this, tx)

    if (!wait) return response
    await response.wait()
    return response
  }

  signAndSend = async (
    accountAddress: string,
    wallet: IStarknetWallet,
    calls: Call[],
    wait = false,
  ) => {
    const account = new Account(this.provider, accountAddress, wallet.signer)

    const tx = await account.execute(calls)
    const response = wrapResponse(this, tx)
    if (!wait) return response

    await response.wait()
    return response
  }
}
