import { TransactionResponse } from '../transaction'
import {
  RpcProvider as StarknetProvider,
  InvokeFunctionResponse,
  DeployContractResponse,
  CompiledContract,
  Account,
  Call,
  constants,
} from 'starknet'
import { IStarknetWallet } from '../wallet'

// TODO: Move to gauntlet-core
interface IProvider<P> {
  provider: P
  send: () => Promise<TransactionResponse>
  declareAndDeployContract: (
    contract: CompiledContract,
    compiledClassHash: string,
    input: any,
    wait?: boolean,
    salt?: number,
  ) => Promise<TransactionResponse>
  deployContract: (
    classHash: string,
    input: any,
    wait?: boolean,
    salt?: number,
  ) => Promise<TransactionResponse>
  declareContract: (
    contract: CompiledContract,
    compiledClassHash?: string,
    wait?: boolean,
  ) => Promise<TransactionResponse>
  signAndSend: (calls: Call[], wait?: boolean) => Promise<TransactionResponse>
}

export interface IStarknetProvider extends IProvider<StarknetProvider> { }
export const makeProvider = (
  url: string,
  wallet?: IStarknetWallet,
): IProvider<StarknetProvider> => {
  return new Provider(url, wallet)
}

export const wrapResponse = (
  provider: IStarknetProvider,
  response: InvokeFunctionResponse | DeployContractResponse,
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
      txResponse.code = status.finality_status // For some reason, starknet does not consider any other status than "TRANSACTION_RECEIVED"
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

  constructor(nodeUrl: string, wallet?: IStarknetWallet) {
    this.provider = new StarknetProvider({ nodeUrl })
    if (wallet) {
      this.account = new Account(
        this.provider,
        wallet.getAccountAddress(),
        wallet.signer,
        /* cairoVersion= */ null, // don't set cairo version so that it's automatically detected from the contract
        /* transactionVersion= */ constants.TRANSACTION_VERSION.V3,
      )
    }
  }

  setAccount(wallet: IStarknetWallet) {
    this.account = new Account(
      this.provider,
      wallet.getAccountAddress(),
      wallet.signer,
      /* cairoVersion= */ null,
      /* transactionVersion= */ constants.TRANSACTION_VERSION.V3,
    )
  }

  send = async () => {
    // Use provider to send tx and wrap it in our type
    return {} as TransactionResponse
  }

  /**
   * Compiles the contract and declares it using the generated ABI.
   * Then deploys an instance of the declared contract.
   * If contract has already been declared it will only be deployed.
   */
  declareAndDeployContract = async (
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

  /**
   * Compiles the contract and declares it using the generated ABI.
   */
  declareContract = async (contract: CompiledContract, compiledClassHash?: string, wait = true) => {
    const tx = await this.account.declare({
      contract,
      compiledClassHash,
    })

    const response = wrapResponse(this, tx, 'not applicable for declares')

    if (!wait) return response
    await response.wait()
    return response
  }

  /**
   * Deploys a contract given a class hash
   */
  deployContract = async (classHash: string, input: any = [], wait = true, salt = undefined) => {
    const tx = await this.account.deployContract({
      classHash: classHash,
      salt: salt ? '0x' + salt.toString(16) : salt,
      ...(!!input && input.length > 0 && { constructorCalldata: input }),
    })
    const response = wrapResponse(this, tx)

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
