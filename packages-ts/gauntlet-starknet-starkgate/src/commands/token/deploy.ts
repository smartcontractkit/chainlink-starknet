import { BN } from '@chainlink/gauntlet-core/dist/utils'
import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  Validation,
} from '@chainlink/gauntlet-starknet'
import { shortString } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'

type UserInput = {
  name: string
  symbol: string
  decimals: string
  minter?: string
}

type ContractInput = [name: string, symbol: string, decimals: string, minter: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  if (flags.link) {
    return {
      name: 'Chainlink LINK Token',
      symbol: 'LINK',
      decimals: '18',
    }
  }

  return {
    name: flags.name,
    symbol: flags.symbol,
    decimals: flags.decimals,
    minter: flags.minter,
  }
}

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountPublicKey()
  return [
    shortString.encodeShortString(input.name),
    shortString.encodeShortString(input.symbol),
    input.decimals,
    input.minter || defaultWallet,
  ]
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to deploy an ERC20 Token Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.TOKEN,
    function: 'deploy',
    examples: [
      `${CATEGORIES.TOKEN}:deploy --network=<NETWORK> --link`,
      `${CATEGORIES.TOKEN}:deploy --network=<NETWORK> --link`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
