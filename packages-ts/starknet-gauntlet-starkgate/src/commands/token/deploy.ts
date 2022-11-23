import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  owner?: string
}

type ContractInput = [owner: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    owner: flags.owner,
  }
}

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountAddress()
  return [input.owner || defaultWallet]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy the LINK Token Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.TOKEN,
  category: CATEGORIES.TOKEN,
  action: 'deploy',
  ux: {
    description: 'Deploys the LINK Token contract',
    examples: [
      `${CATEGORIES.TOKEN}:deploy --network=<NETWORK>`,
      `${CATEGORIES.TOKEN}:deploy --network=<NETWORK> --owner=<ACCOUNT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
