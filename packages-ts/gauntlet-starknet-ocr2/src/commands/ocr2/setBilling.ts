import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader } from '../../lib/contracts'

type UserInput = {
  observationPaymentGjuels: number
  transmissionPaymentGjuels: number
}

type ContractInput = [
  {
    observation_payment_gjuels: number
    transmission_payment_gjuels: number
  },
]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    observationPaymentGjuels: flags.observationPaymentGjuels,
    transmissionPaymentGjuels: flags.transmissionPaymentGjuels,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [
    {
      observation_payment_gjuels: new BN(input.observationPaymentGjuels).toNumber(),
      transmission_payment_gjuels: new BN(input.transmissionPaymentGjuels).toNumber(),
    },
  ]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  let observationPayment: BN
  let transmissionPayment: BN

  try {
    observationPayment = new BN(input.observationPaymentGjuels)
    transmissionPayment = new BN(input.transmissionPaymentGjuels) // parse as integers
  } catch {
    throw new Error(
      `observationPaymentGjuels=${input.observationPaymentGjuels} and ` +
        `transmissionPaymentGjuels=${input.transmissionPaymentGjuels} must both be integers`,
    )
  }
  if (observationPayment.isNeg() || transmissionPayment.isNeg()) {
    throw new Error(
      `observationPaymentGjuels=${input.observationPaymentGjuels} and ` +
        `transmissionPaymentGjuels=${input.transmissionPaymentGjuels} cannot be negative`,
    )
  }
  return true
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'set_billing',
    examples: [
      `${CATEGORIES.OCR2}:set_billing --network=<NETWORK> --observationPaymentGjuels=<AMOUNT> --transmissionPaymentGjuels=<AMOUNT> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
