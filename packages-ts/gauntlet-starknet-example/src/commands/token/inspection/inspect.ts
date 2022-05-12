import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { InspectCommandConfig, IStarknetProvider, makeInspectionCommand } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../../lib/categories'
import { tokenContract } from '../../../lib/contracts'

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
    category: CATEGORIES.TOKEN,
    function: 'inspect',
    examples: ['yarn gauntlet token:inspect --network=<NETWORK>'],
  },
  queries: ['get_balance'],
  makeComparisionData,
  contract: tokenContract,
}

export default makeInspectionCommand(commandConfig)
