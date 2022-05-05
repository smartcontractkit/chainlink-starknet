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

class Provider implements IStarknetProvider {
  provider: StarknetProvider

  constructor(baseUrl: string) {
    this.provider = new StarknetProvider({ baseUrl })
  }

  private wrapResponse = (response: AddTransactionResponse): TransactionResponse => {
    return {
      hash: response.transaction_hash,
      address: response.address,
      wait: async () => ({
        success: (await this.provider.waitForTransaction(response.transaction_hash)) === undefined,
      }),
      tx: response,
    }
  }

  send = async () => {
    // Use provider to send tx and wrap it in our type
    return {} as TransactionResponse
  }

  deployContract = async (contract: CompiledContract, wait = true) => {
    const tx = await this.provider.deployContract({
      contract,
    })

    const response = this.wrapResponse(tx)

    if (!wait) return response
    await response.wait()
    return response
  }
}
