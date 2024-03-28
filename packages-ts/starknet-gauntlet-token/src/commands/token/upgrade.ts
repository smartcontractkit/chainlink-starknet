import { makeExecuteCommand, upgradeCommandConfig } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader } from '../../lib/contracts'

export default makeExecuteCommand(
  upgradeCommandConfig(CATEGORIES.TOKEN, CATEGORIES.TOKEN, tokenContractLoader),
)
