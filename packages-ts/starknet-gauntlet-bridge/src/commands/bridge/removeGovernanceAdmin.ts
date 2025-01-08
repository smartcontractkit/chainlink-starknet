import { makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { createRemoveAdminCommandConfig } from '../../lib/createAdminCommand'

const commandConfig = createRemoveAdminCommandConfig('remove_governance_admin')

export default makeExecuteCommand(commandConfig)
