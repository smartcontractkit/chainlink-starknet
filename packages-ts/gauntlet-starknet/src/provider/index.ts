import { TransactionResponse } from '../transaction'
import { Provider as StarknetProvider, AddTransactionResponse, CompiledContract } from 'starknet'

// TODO: Move to gauntlet-core
interface IProvider<P> {
  provider: P
  send: () => Promise<TransactionResponse>
  deployContract: (contract: CompiledContract, wait?: boolean) => Promise<TransactionResponse>
}

export interface IStarknetProvider extends IProvider<StarknetProvider> {}

export const makeProvider = (url: string): IProvider<StarknetProvider> => {
  return new Provider(url)
}

export const wrapResponse = (
  provider: IStarknetProvider,
  response: AddTransactionResponse,
  address?: string,
): TransactionResponse => {
  const txResponse = {
    hash: response.transaction_hash,
    address: address || response.address,
    wait: async () => {
      // Success if does not throw
      const success = (await provider.provider.waitForTransaction(response.transaction_hash)) === undefined
      const status = await provider.provider.getTransactionStatus(response.transaction_hash)
      txResponse.tx.code = status.tx_status as any // For some reason, starknet does not consider any other status than "TRANSACTION_RECEIVED"
      return {
        success,
      }
    },
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

  deployContract = async (contract: CompiledContract, wait = true) => {
    const tx = await this.provider.deployContract({
      contract,
    })

    const response = wrapResponse(this, tx)

    if (!wait) return response
    await response.wait()
    return response
  }
}
