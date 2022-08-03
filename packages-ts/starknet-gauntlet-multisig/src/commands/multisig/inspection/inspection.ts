import { InspectCommandConfig, IStarknetProvider, makeInspectionCommand } from '@chainlink/starknet-gauntlet'
import { toBN, toHex } from 'starknet/dist/utils/number'
import { CATEGORIES } from '../../../lib/categories'
import { contractLoader } from '../../../lib/contracts'

type QueryResult = {
  signers: string[]
  threshold: number
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const [signers, threshold] = results
  return {
    toCompare: null,
    result: {
      signers: signers.signers.map((o) => toHex(o)),
      threshold: toBN(threshold.confirmations_required).toNumber(),
    },
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.MULTISIG,
    function: 'inspect',
    examples: [`${CATEGORIES.MULTISIG}:inspect --network=<NETWORK>`],
  },
  queries: ['get_signers', 'get_confirmations_required'],
  makeComparisionData,
  loadContract: contractLoader,
}

export default makeInspectionCommand(commandConfig)
