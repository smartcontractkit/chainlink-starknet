import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  ExecuteCommandInstance,
  CommandCtor,
  makeWallet,
  makeProvider,
  Dependencies,
  Env,
} from '@chainlink/starknet-gauntlet'

import { L1Commands, L2Commands } from './commands'

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

// const registeredCommands = Commands.map(registerExecuteCommand)

export { L1Commands, L2Commands }
// export default [...registeredCommands]
