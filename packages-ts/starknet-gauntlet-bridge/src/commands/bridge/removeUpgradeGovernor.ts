import { makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { createRemoveAdminCommandConfig } from '../../lib/createAdminCommand'

const commandConfig = createRemoveAdminCommandConfig('remove_upgrade_governor')

export default makeExecuteCommand(commandConfig)
