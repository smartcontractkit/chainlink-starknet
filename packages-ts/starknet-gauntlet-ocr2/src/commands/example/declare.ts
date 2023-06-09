import { makeExecuteCommand, declareCommandConfig } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { CONTRACT_LIST, aggregatorConsumerLoader } from '../../lib/contracts'

export default makeExecuteCommand(
  declareCommandConfig(
    CONTRACT_LIST.AGGREGATOR_CONSUMER,
    CATEGORIES.OCR2,
    aggregatorConsumerLoader,
  ),
)
