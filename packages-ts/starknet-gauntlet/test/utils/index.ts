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
} from '../../src/index'

export { startNetwork, IntegratedDevnet } from './network'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../__mocks__/${name}.json`).toString('ascii'))
}

export const loadExampleContract = (): any => {
  return { contract: loadContract('example') }
}

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

export const TIMEOUT = 900000
export const LOCAL_URL = 'http://127.0.0.1:5050/'
export const devnetPrivateKey = '0xe3e70682c2094cac629f6fbed82c07cd'
export const devnetAccount0Address =
  '0x7e00d496e324876bbc8531f2d9a82bf154d1a04a50218ee74cdd372f75a551a'
// export const devnetKeyPair = ec.getKeyPair(devnetPrivateKey);

export const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: (flags) => {
      return {
        providerUrl: LOCAL_URL,
        pk: (flags.pk as string) || devnetPrivateKey,
        account: (flags.account as string) || devnetAccount0Address,
      }
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

export const registerInspectCommand = <QueryResult>(
  registerCommand: (
    deps: Omit<Dependencies, 'makeWallet'>,
  ) => CommandCtor<InspectCommandInstance<QueryResult>>,
) => {
  const deps: Omit<Dependencies, 'makeWallet'> = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: (flags) => {
      return {
        providerUrl: LOCAL_URL,
        pk: (flags.pk as string) || devnetPrivateKey,
        account: (flags.account as string) || devnetAccount0Address,
      }
    },
    makeProvider: makeProvider,
  }
  return registerCommand(deps)
}
