import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  ExecuteCommandInstance,
  CommandCtor,
  makeWallet,
  makeProvider,
  Dependencies,
  Env,
} from '@chainlink/gauntlet-starknet'

import { executeCommands, inspectionCommands } from './commands'
import { wrapCommand } from './wrapper'

const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: logger,
    prompt: prompt,
    makeEnv: (flags) => {
      const env: Env = {
        providerUrl: process.env.NODE_URL || 'https://alpha4.starknet.io',
        pk: process.env.PRIVATE_KEY,
        account: process.env.ACCOUNT,
      }
      return env
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

const registeredCommands = executeCommands.map(registerExecuteCommand)

export { executeCommands, inspectionCommands, wrapCommand }
export default [...registeredCommands]
