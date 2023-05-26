import { CONTRACT_TYPES, ExecuteCommandConfig, ExecutionContext, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetBilling, SetBillingInput } from '@chainlink/gauntlet-contracts-ocr2'

type StarknetSetBillingInput = SetBillingInput & { gasBase: number; gasPerSignature: number }

type ContractInput = [
  {
    observation_payment_gjuels: number
    transmission_payment_gjuels: number
    gas_base: number
    gas_per_signature: number
  },
]

const makeContractInput = async (input: StarknetSetBillingInput, ctx: ExecutionContext): Promise<ContractInput> => {
  if (ctx.rdd) {
    const contract = ctx.rdd[CONTRACT_TYPES.AGGREGATOR][ctx.contractAddress];
    input = contract.billing;
  }

  return [
    {
      observation_payment_gjuels: input.observationPaymentGjuels,
      transmission_payment_gjuels: input.transmissionPaymentGjuels,
      gas_base: input.gasBase,
      gas_per_signature: input.gasPerSignature,
    },
  ]
}

const commandConfig: ExecuteCommandConfig<StarknetSetBillingInput, ContractInput> = {
  ...SetBilling,
  makeUserInput: (flags: any, args: any): StarknetSetBillingInput => {
    if (flags.input) return flags.input as StarknetSetBillingInput
    return {
      observationPaymentGjuels: parseInt(flags.observationPaymentGjuels),
      transmissionPaymentGjuels: parseInt(flags.transmissionPaymentGjuels),
      gasBase: parseInt(flags.gasBase),
      gasPerSignature: parseInt(flags.gasPerSignature),
    }
  },
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
