import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
  InspectUserInput,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { l2BridgeContractLoader } from '../../lib/contracts'
import { validateAndParseAddress } from 'starknet'

type UserInput = {}

type ContractInput = null

type QueryResult = {
  governor: string
  l1Bridge: string
  l2Token: string
}

const makeUserInput = async (flags, args): Promise<InspectUserInput<UserInput, null>> => {
  if (flags.input) return flags.input as InspectUserInput<UserInput, null>

  return {
    input: {},
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput[]> => {
  return []
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const [governor, l1Bridge, l2Token] = results

  return {
    toCompare: null,
    result: {
      governor: validateAndParseAddress(governor[0]),
      l1Bridge: validateAndParseAddress(l1Bridge[0]),
      l2Token: validateAndParseAddress(l2Token[0]),
    },
  }
}

const commandConfig: InspectCommandConfig<UserInput, ContractInput, null, QueryResult> = {
  ux: {
    category: CATEGORIES.L2_BRIDGE,
    function: 'inspect',
    examples: [`${CATEGORIES.L2_BRIDGE}:inspect --network=<NETWORK> <L2_BRIDGE_ADDRESS>`],
  },
  queries: ['get_governor', 'get_l1_bridge', 'get_l2_token'],
  makeUserInput,
  makeContractInput,
  makeComparisionData,
  loadContract: l2BridgeContractLoader,
}

export default makeInspectionCommand(commandConfig)
