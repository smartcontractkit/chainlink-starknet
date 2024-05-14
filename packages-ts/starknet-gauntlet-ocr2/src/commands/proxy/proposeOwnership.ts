import { makeExecuteCommand, proposeOwnershipCommandConfig } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ProxyLoader } from '../../lib/contracts'

export default makeExecuteCommand(
  proposeOwnershipCommandConfig(CATEGORIES.PROXY, CATEGORIES.PROXY, ocr2ProxyLoader),
)
