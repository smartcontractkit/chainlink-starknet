import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'
import { isValidAddress } from '@chainlink/starknet-gauntlet'

export interface UserInput {
  starkNetMessaging: number
  configAC: string
  gasPriceL1Feed: string
  source: string
  gasEstimate: number
  l2Feed: string
}

type ContractInput = [
  starkNetMessaging: number,
  configAC: string,
  gasPriceL1Feed: string,
  source: string,
  l2Feed: string,
  gasEstimate: number,
]

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [
    input.starkNetMessaging,
    input.configAC,
    input.gasPriceL1Feed,
    input.source,
    input.l2Feed,
    input.gasEstimate,
  ]
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    starkNetMessaging: flags.starkNetMessaging,
    configAC: flags.configAC,
    gasPriceL1Feed: flags.gasPriceL1Feed,
    source: flags.source,
    gasEstimate: flags.gasEstimate,
    l2Feed: flags.l2Feed,
  }
}

const validateStarkNetMessaging = async (input) => {
  if (!isValidAddress(input.starkNetMessaging)) {
    throw new Error(`Invalid starkNetMessaging Address: ${input.starkNetMessaging}`)
  }
  return true
}

const validateConfigAC = async (input) => {
  if (!isValidAddress(input.configAC)) {
    throw new Error(`Invalid configAC Address: ${input.configAC}`)
  }
  return true
}

const validateGasPriceL1Feed = async (input) => {
  if (!isValidAddress(input.gasPriceL1Feed)) {
    throw new Error(`Invalid gasPriceL1Feed Address: ${input.gasPriceL1Feed}`)
  }
  return true
}

const validateSourceAggregator = async (input) => {
  if (!isValidAddress(input.source)) {
    throw new Error(`Invalid source Address: ${input.source}`)
  }
  return true
}

const validateGasEstimate = async (input) => {
  if (isNaN(Number(input.gasEstimate))) {
    throw new Error(`Invalid gasEstimate (must be number): ${input.gasEstimate}`)
  }
  return true
}

const validateL2Feed = async (input) => {
  if (!isValidAddress(input.l2Feed)) {
    throw new Error(`Invalid l2Feed Address: ${input.l2Feed}`)
  }
  return true
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'deploy',
  ux: {
    description:
      'Deploys a StarknetValidator contract. Starknet messaging contract is address officially deployed by starkware industries. ',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:deploy --starkNetMessaging=<ADDRESS> --configAC=<ADDRESS>--gasPriceL1Feed=<ADDRESS> --source=<ADDRESS> --gasEstimate=<GAS_ESTIMATE> --l2Feed=<ADDRESS> --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [
    validateStarkNetMessaging,
    validateConfigAC,
    validateGasPriceL1Feed,
    validateGasEstimate,
    validateSourceAggregator,
    validateL2Feed,
  ],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
