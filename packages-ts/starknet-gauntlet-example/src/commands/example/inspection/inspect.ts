import { BN } from '@chainlink/gauntlet-core/dist/utils'
import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../../lib/categories'
import { tokenContractLoader } from '../../../lib/contracts'

type QueryResult = {
  balance: string
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const [balance] = results
  return {
    toCompare: null,
    result: {
      balance: new BN(balance.res).toString(),
    },
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.EXAMPLE,
    function: 'inspect',
    examples: ['yarn gauntlet token:inspect --network=<NETWORK>'],
  },
  queries: ['get_balance'],
  makeComparisionData,
  loadContract: tokenContractLoader,
}

export default makeInspectionCommand(commandConfig)
