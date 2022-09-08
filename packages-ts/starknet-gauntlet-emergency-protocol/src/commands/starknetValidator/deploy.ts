import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export interface UserInput {
  starkNetMessaging: number
  l2UptimeFeedAddr: string
}

type ContractInput = [starkNetMessaging: number, l2UptimeFeedAddr: string]

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.starkNetMessaging, input.l2UptimeFeedAddr]
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    l2UptimeFeedAddr: flags.l2UptimeFeedAddr,
    starkNetMessaging: flags.starkNetMessaging,
  }
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'deploy',
  ux: {
    description: 'Deploys a StarknetValidator contract',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:deploy --starkNetMessaging=0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4 --l2UptimeFeedAddr=0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
