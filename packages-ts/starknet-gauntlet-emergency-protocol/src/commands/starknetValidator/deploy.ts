import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

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

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'deploy',
  ux: {
    description: 'Deploys a StarknetValidator contract',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:deploy --starkNetMessaging=0xde29d060D45901Fb19ED6C6e959EB22d8626708e --configAC=0x42f4802128C56740D77824046bb13E6a38874331 --gasPriceL1Feed=0xdcb95Cd00d32d02b5689CE020Ed67f4f91ee5942 --source=0x42f4802128C56740D77824046bb13E6a38874331 --gasEstimate=0 --l2Feed=0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
