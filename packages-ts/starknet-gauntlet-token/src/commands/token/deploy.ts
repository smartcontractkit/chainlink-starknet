import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  minter?: string
  owner?: string
  classHash?: string
}

type ContractInput = [minter: string, owner: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    minter: flags.minter,
    owner: flags.owner,
    classHash: flags.classHash
  }
}

const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash) || input.classHash === undefined) {
    return true
  }
  throw new Error(`Invalid Class Hash: ${input.classHash}`)
}

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountAddress()
  return [input.minter || defaultWallet, input.owner || defaultWallet]
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
      `${CATEGORIES.TOKEN}:deploy --network=<NETWORK> --owner=<ACCOUNT_ADDRESS> --classHash=<CLASS_HASH>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateClassHash],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
