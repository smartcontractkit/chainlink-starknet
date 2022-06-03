import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetBilling, SetBillingInput } from '@chainlink/gauntlet-contracts-ocr2'

type ContractInput = [
  {
    observation_payment_gjuels: number
    transmission_payment_gjuels: number
  },
]

const makeContractInput = async (input: SetBillingInput): Promise<ContractInput> => {
  return [
    {
      observation_payment_gjuels: new BN(input.observationPaymentGjuels).toNumber(),
      transmission_payment_gjuels: new BN(input.transmissionPaymentGjuels).toNumber(),
    },
  ]
}

const commandConfig: ExecuteCommandConfig<SetBillingInput, ContractInput> = {
  ...SetBilling,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
