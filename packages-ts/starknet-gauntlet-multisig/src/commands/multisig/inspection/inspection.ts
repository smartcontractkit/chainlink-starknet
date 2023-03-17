import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
} from '@chainlink/starknet-gauntlet'
import { num } from 'starknet'
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
      signers: signers.signers.map((o) => num.toHex(o)),
      threshold: threshold.confirmations_required,
    },
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.MULTISIG,
    function: 'inspect',
    examples: [`${CATEGORIES.MULTISIG}:inspect --network=<NETWORK>`],
  },
  queries: ['get_signers', 'get_threshold'],
  makeComparisionData,
  loadContract: contractLoader,
}

export default makeInspectionCommand(commandConfig)
