import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { isValidAddress } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { l2BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  governor?: string
}

type ContractInput = [governor: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    governor: flags.governor,
  }
}

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountAddress()
  return [input.governor || defaultWallet]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!isValidAddress(input.governor)) {
    throw new Error(`Invalid governor: ${input.governor}`)
  }

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy an L2 Bridge Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L2_BRIDGE,
  category: CATEGORIES.L2_BRIDGE,
  action: 'deploy',
  ux: {
    description: 'Deploys an L2 token bridge',
    examples: [
      `${CATEGORIES.L2_BRIDGE}:deploy --network=<NETWORK> --governor=[OWNER_ADDRESS]`,
      `${CATEGORIES.L2_BRIDGE}:deploy --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l2BridgeContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
