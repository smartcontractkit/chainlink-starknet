import { BN } from '@chainlink/gauntlet-core/dist/utils'
import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  Validation,
} from '@chainlink/starknet-gauntlet'
import { shortString } from 'starknet'
import { isContext } from 'vm'
import { CATEGORIES } from '../../lib/categories'
import { bridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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
  deps.logger.info(`About to deploy an L2 Bridge Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.BRIDGE,
  category: CATEGORIES.BRIDGE,
  action: 'deploy',
  ux: {
    description: 'Deploys an L2 token bridge',
    examples: [`${CATEGORIES.BRIDGE}:deploy --network=<NETWORK> --governor=[ADDRESS]`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: bridgeContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
