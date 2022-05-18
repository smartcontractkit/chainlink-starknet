import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  ExecuteCommandInstance,
  CommandCtor,
  makeWallet,
  makeProvider,
  Dependencies,
} from '@chainlink/gauntlet-starknet'

import Commands from './commands'

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

const registeredCommands = Commands.map(registerExecuteCommand)

export { Commands }
export default [...registeredCommands]
