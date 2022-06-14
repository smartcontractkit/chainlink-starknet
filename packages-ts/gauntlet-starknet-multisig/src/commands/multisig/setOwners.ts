import { BeforeExecute, ExecuteCommandConfig, ExecutionContext, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'
import { validateOwners } from './deploy'

type UserInput = {
  owners: string[]
}

type ContractInput = [owners: string[]]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    owners: flags.owners,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owners.map((addr) => number.toFelt(number.toBN(addr)))]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger
    .info(`About to set owners on the Multisig Contract (${context.contractAddress}) with the following details:
    - New Owners: ${input.user.owners}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    description: 'Set owners of the multisig account',
    examples: [`${CATEGORIES.MULTISIG}:set_owners --network=<NETWORK> --owners=[OWNERS_LIST]`],
  },
  category: CATEGORIES.MULTISIG,
  action: 'set_owners',
  contractId: CATEGORIES.MULTISIG,
  makeUserInput,
  makeContractInput,
  validations: [validateOwners],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
