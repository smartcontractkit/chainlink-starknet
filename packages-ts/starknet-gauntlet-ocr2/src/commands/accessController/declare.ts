import { makeExecuteCommand, declareCommandConfig } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { accessControllerContractLoader } from '../../lib/contracts'

export default makeExecuteCommand(
  declareCommandConfig(
    CATEGORIES.ACCESS_CONTROLLER,
    CATEGORIES.ACCESS_CONTROLLER,
    accessControllerContractLoader,
  ),
)
