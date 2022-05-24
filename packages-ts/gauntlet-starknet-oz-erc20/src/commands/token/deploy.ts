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
  initialSupply: string
  recipient?: string
  owner?: string
}

type ContractInput = [
  name: string,
  symbol: string,
  decimals: string,
  initial_supply: string,
  recipient: string,
  owner: string,
]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  if (flags.link) {
    return {
      name: 'Chainlink LINK Token',
      symbol: 'LINK',
      decimals: '18',
      initialSupply: '10000000',
    }
  }

  return {
    name: flags.name,
    symbol: flags.symbol,
    decimals: flags.decimals,
    initialSupply: flags.initialSupply,
    recipient: flags.recipient,
    owner: flags.owner,
  }
}

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  const defaultWallet = context.wallet.getAccountPublicKey()
  return [
    shortString.encodeShortString(input.name),
    shortString.encodeShortString(input.symbol),
    input.decimals,
    new BN(input.initialSupply).toString(),
    input.recipient || defaultWallet,
    input.owner || defaultWallet,
  ]
}

const validate: Validation<UserInput> = async (input) => {
  // todo: validate every fiels exists
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
