import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
  InspectUserInput,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader } from '../../lib/contracts'
import { uint256ToBN } from 'starknet/dist/utils/uint256'

type UserInput = {
  account: string
}

type ContractInput = string

type QueryResult = {
  balance: string
}

const makeUserInput = async (flags, args): Promise<InspectUserInput<UserInput, null>> => {
  if (flags.input) return flags.input as InspectUserInput<UserInput, null>

  return {
    input: {
      account: flags.account,
    },
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput[]> => {
  if (!isValidAddress(input.account)) throw new Error(`Invalid account address: ${input.account}`)
  return [input.account]
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const [queryRes] = results
  const [balance] = queryRes

  return {
    toCompare: null,
    result: {
      balance: uint256ToBN(balance).toString(),
    },
  }
}

const commandConfig: InspectCommandConfig<UserInput, ContractInput, null, QueryResult> = {
  ux: {
    category: CATEGORIES.TOKEN,
    function: 'balance_of',
    examples: [
      'yarn gauntlet token:balanceOf --network=<NETWORK> --account=<ACCOUNT> <CONTRACT_ADDRESS>',
    ],
  },
  queries: ['balanceOf'],
  makeUserInput,
  makeContractInput,
  makeComparisionData,
  loadContract: tokenContractLoader,
}

export default makeInspectionCommand(commandConfig)
