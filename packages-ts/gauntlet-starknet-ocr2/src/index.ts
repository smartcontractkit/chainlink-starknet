import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  ExecuteCommandInstance,
  CommandCtor,
  makeWallet,
  makeProvider,
  Dependencies,
} from '@chainlink/starknet-gauntlet'

import { executeCommands } from './commands'

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
        publicKey: process.env.PUBLIC_KEY,
        account: process.env.ACCOUNT,
        billingAccessController: process.env.BILLING_ACCESS_CONTROLLER,
        link: process.env.LINK,
      }
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

const registeredCommands = executeCommands.map(registerExecuteCommand)

export { executeCommands }
export default [...registeredCommands]
