import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { InspectCommandConfig, IStarknetProvider, makeInspectionCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../../lib/categories'
import { ocr2ContractLoader } from '../../../lib/contracts'

type QueryResult = {}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const tx = await provider.provider.getTransactionReceipt(
    '0x475c15d6836972234c0542044fce7784cc61e8c5654d050aacadb918d8f3021',
  )
  console.log(tx.events)
  return {
    toCompare: null,
    result: {},
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'inspect',
    examples: ['yarn gauntlet ocr2:inspect --network=<NETWORK>'],
  },
  queries: [],
  makeComparisionData,
  loadContract: ocr2ContractLoader,
}

export default makeInspectionCommand(commandConfig)
