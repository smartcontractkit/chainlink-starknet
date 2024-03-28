import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { accessControllerContractLoader, CONTRACT_LIST } from '../../lib/contracts'
import { validateClassHash } from '../../lib/utils'

type UserInput = {
  owner: string
  classHash?: string
}

type ContractInput = [owner: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
    classHash: flags.classHash,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owner]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCESS_CONTROLLER,
  category: CATEGORIES.ACCESS_CONTROLLER,
  action: 'deploy',
  ux: {
    description: 'Deploys an Access Controller Contract',
    examples: [
      `${CATEGORIES.ACCESS_CONTROLLER}:deploy --network=<NETWORK> --classHash=<CLASS_HASH>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateClassHash],
  loadContract: accessControllerContractLoader,
}

export default makeExecuteCommand(commandConfig)
