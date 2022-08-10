import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet' // todo: use @chainlink/evm-gauntlet
import { CATEGORIES } from '../../lib/categories'
import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountPublicKey()
  return [input.governor || defaultWallet]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to deploy L1 Bridge Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE,
  category: CATEGORIES.L1_BRIDGE,
  action: 'deploy',
  ux: {
    description: 'Deploys an L1 token bridge',
    examples: [`${CATEGORIES.L1_BRIDGE}:deploy --network=<NETWORK> --governor=[ADDRESS]`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: l1BridgeContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
