import { makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { createRegisterAdminCommandConfig } from '../../lib/createAdminCommand'

const commandConfig = createRegisterAdminCommandConfig('register_governance_admin')

export default makeExecuteCommand(commandConfig)
