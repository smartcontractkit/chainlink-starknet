import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
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

const makeContractInput = async (input: StarknetSetBillingInput): Promise<ContractInput> => {
  return [
    {
      observation_payment_gjuels: new BN(input.observationPaymentGjuels).toNumber(),
      transmission_payment_gjuels: new BN(input.transmissionPaymentGjuels).toNumber(),
      gas_base: new BN(input.gasBase).toNumber(),
      gas_per_signature: new BN(input.gasPerSignature).toNumber(),
    },
  ]
}

const commandConfig: ExecuteCommandConfig<StarknetSetBillingInput, ContractInput> = {
  ...SetBilling,
  makeUserInput: (flags: any, args: any): StarknetSetBillingInput => {
    if (flags.input) return flags.input as StarknetSetBillingInput
    return {
      observationPaymentGjuels: flags.observationPaymentGjuels,
      transmissionPaymentGjuels: flags.transmissionPaymentGjuels,
      gasBase: flags.gasBase,
      gasPerSignature: flags.gasPerSignature,
    }
  },
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
