import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  ExecuteCommandInstance,
  CommandCtor,
  makeWallet,
  makeProvider,
  Dependencies,
} from '@chainlink/starknet-gauntlet'

import { executeCommands, inspectionCommands } from './commands'

const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: logger,
    prompt: prompt,
    makeEnv: (flags) => {
      return {
        providerUrl: process.env.NODE_URL || 'https://alpha4.starknet.io',
        pk: process.env.PRIVATE_KEY,
        account: process.env.ACCOUNT,
      }
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

const registeredCommands = executeCommands.map(registerExecuteCommand)

export { executeCommands, inspectionCommands }
export default [...registeredCommands]
