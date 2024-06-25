import {
  CONTRACT_TYPES,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetBilling, SetBillingInput } from '@chainlink/gauntlet-contracts-ocr2'
import { getRDD } from '@chainlink/starknet-gauntlet'

type StarknetSetBillingInput = SetBillingInput & { gasBase: number; gasPerSignature: number }

type ContractInput = [
  {
    observation_payment_gjuels: number
    transmission_payment_gjuels: number
    gas_base: number
    gas_per_signature: number
  },
]

const makeContractInput = async (
  input: StarknetSetBillingInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
  return [
    {
      observation_payment_gjuels: input.observationPaymentGjuels,
      transmission_payment_gjuels: input.transmissionPaymentGjuels,
      gas_base: input.gasBase || 0,
      gas_per_signature: input.gasPerSignature || 0,
    },
  ]
}

const commandConfig: ExecuteCommandConfig<StarknetSetBillingInput, ContractInput> = {
  ...SetBilling,
  makeUserInput: (flags: any, args: any): StarknetSetBillingInput => {
    if (flags.input) return flags.input as StarknetSetBillingInput
    if (flags.rdd) {
      const rdd = getRDD(flags.rdd)
      const contractAddr = args[0]
      const contract = rdd[CONTRACT_TYPES.AGGREGATOR][contractAddr]
      return contract.billing
    }

    return {
      observationPaymentGjuels: parseInt(flags.observationPaymentGjuels),
      transmissionPaymentGjuels: parseInt(flags.transmissionPaymentGjuels),
      gasBase: parseInt(flags.gasBase || '0'), // optional
      gasPerSignature: parseInt(flags.gasPerSignature || '0'), //optional
    }
  },
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
