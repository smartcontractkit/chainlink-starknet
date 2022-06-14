import { CompiledContract, json } from 'starknet'
import fs from 'fs'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  CommandCtor,
  Dependencies,
  ExecuteCommandInstance,
  InspectCommandInstance,
  makeProvider,
  makeWallet,
} from '@chainlink/gauntlet-starknet'

export { startNetwork, IntegratedDevnet } from './network'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../__mocks__/${name}.json`).toString('ascii'))
}

export const loadExampleContract = () => loadContract('example')

export const noop = () => {}

export const noopLogger: typeof logger = {
  table: noop,
  log: noop,
  info: noop,
  warn: noop,
  success: noop,
  error: noop,
  loading: noop,
  line: noop,
  style: () => '',
  debug: noop,
  time: noop,
}

export const noopPrompt: typeof prompt = async () => {}

export const TIMEOUT = 100000
export const LOCAL_URL = 'http://127.0.0.1:5050/'

export const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: (flags) => {
      return {
        providerUrl: LOCAL_URL,
        pk: flags.pk,
        account: flags.account,
      }
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

export const registerInspectCommand = <QueryResult>(
  registerCommand: (deps: Omit<Dependencies, 'makeWallet'>) => CommandCtor<InspectCommandInstance<QueryResult>>,
) => {
  const deps: Omit<Dependencies, 'makeWallet'> = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: (flags) => {
      return {
        providerUrl: LOCAL_URL,
        pk: flags.pk,
        account: flags.account,
      }
    },
    makeProvider: makeProvider,
  }
  return registerCommand(deps)
}
