import { TransactionResponse } from '../transaction'
import {
  RpcProvider as StarknetProvider,
  DeclareContractResponse,
  InvokeFunctionResponse,
  DeployContractResponse,
  CompiledContract,
  Account,
  Call,
  constants,
  UniversalDetails,
  ResourceBounds,
  EstimateFeeResponse,
  ETransactionVersion,
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
  deployAccountContract: (
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

export interface IStarknetProvider extends IProvider<StarknetProvider> {}
export const makeProvider = (
  url: string,
  wallet?: IStarknetWallet,
): IProvider<StarknetProvider> => {
  return new Provider(url, wallet)
}

export const wrapResponse = (
  provider: IStarknetProvider,
  response: InvokeFunctionResponse | DeployContractResponse | DeclareContractResponse,
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
    this.provider = new StarknetProvider({ nodeUrl, specVersion: '0.8' })
    if (wallet) {
      this.account = new Account(
        this.provider,
        wallet.getAccountAddress(),
        wallet.signer,
        undefined,
        ETransactionVersion.V3,
      )
    }
  }

  setAccount(wallet: IStarknetWallet) {
    this.account = new Account(
      this.provider,
      wallet.getAccountAddress(),
      wallet.signer,
      undefined,
      ETransactionVersion.V3,
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
    const tx = await this.account.declareAndDeploy(
      {
        contract,
        compiledClassHash,
        salt: !isNaN(salt) ? '0x' + salt.toString(16) : salt, // convert number to hex or leave undefined
        // unique: false,
        ...(!!input && input.length > 0 && { constructorCalldata: input }),
      },
      {
        version: ETransactionVersion.V3,
      },
    )
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
      salt: !isNaN(salt) ? '0x' + salt.toString(16) : salt,
      ...(!!input && input.length > 0 && { constructorCalldata: input }),
    })
    const response = wrapResponse(this, tx)

    if (!wait) return response
    await response.wait()
    return response
  }

  /**
   * Deploys an account contract using DEPLOY_ACCOUNT given a class hash
   */

  deployAccountContract = async (classHash: string, input: any = [], wait = true, salt = 0) => {
    const tx = await this.account.deployAccount({
      classHash: classHash,
      constructorCalldata: input,
      addressSalt: '0x' + salt.toString(16),
    })
    const response = wrapResponse(this, tx)

    if (!wait) return response
    await response.wait()
    return response
  }

  signAndSend = async (calls: Call[], wait = false) => {
    let feeEstimate
    console.log(calls)
    try {
      feeEstimate = await this.account.estimateFee(calls, { version: ETransactionVersion.V3 })
    } catch (error) {
      console.error('Failed to estimate fee:', error)
      throw error // optionally rethrow if you want the function to still fail
    }
    const resouceBounds = this.feeEstimateToResourceBoundsMapping(feeEstimate)
    const tx = await this.account.execute(calls, { resourceBounds: resouceBounds })
    const response = wrapResponse(this, tx)
    if (!wait) return response

    await response.wait()
    return response
  }

  feeEstimateToResourceBoundsMapping = (
    estimate: EstimateFeeResponse,
    bufferPercent: number = 20, // optional buffer to avoid underestimation
  ): ResourceBounds => {
    const bufferMultiplier = BigInt(100 + bufferPercent)

    const withBuffer = (amount: bigint) => (amount * bufferMultiplier) / 100n

    return {
      l1_gas: {
        max_amount: String(withBuffer(estimate.l1_gas_consumed * estimate.l1_gas_price)),
        max_price_per_unit: String(estimate.l1_gas_price),
      },
      l1_data_gas: {
        max_amount: String(withBuffer(estimate.l1_data_gas_consumed * estimate.l1_data_gas_price)),
        max_price_per_unit: String(estimate.l1_data_gas_price),
      },
      l2_gas: {
        max_amount: String(
          estimate.l2_gas_consumed && estimate.l2_gas_price
            ? withBuffer(estimate.l2_gas_consumed * estimate.l2_gas_price)
            : 0n,
        ),
        max_price_per_unit: String(estimate.l2_gas_price ?? 0n),
      },
    }
  }
}
