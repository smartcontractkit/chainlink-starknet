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
  LoadContractResult,
} from '../../src/index'

export const loadContract = (name: string): LoadContractResult => {
  return {
    contract: json.parse(
      fs.readFileSync(`${__dirname}/../__mocks__/${name}.json`).toString('ascii'),
    ),
    casm: json.parse(
      fs.readFileSync(`${__dirname}/../__mocks__/${name}_compiled.json`).toString('ascii'),
    ),
  }
}

export const loadExampleContract = () => {
  return loadContract('example')
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

export const TIMEOUT = 1800000 // 30 minutes - increased for complex multisig operations
export const LOCAL_URL = 'http://127.0.0.1:5050/'

export type StarknetAccount = Awaited<ReturnType<typeof fetchAccount>>

type PredeployedAccount = {
  address: string
  private_key: string
  initial_balance: string
}

const fetchPredeployedAccounts = async (): Promise<PredeployedAccount[]> => {
  const response = await fetch(`${LOCAL_URL}rpc`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'devnet_getPredeployedAccounts',
      params: {},
    }),
  })
  const data = (await response.json()) as {
    result?: PredeployedAccount[]
    error?: { message: string }
  }
  if (data.error) {
    throw new Error(data.error.message)
  }
  if (!data.result) {
    throw new Error('devnet_getPredeployedAccounts returned no accounts')
  }
  return data.result
}

export const fetchAccount = async (accountIndex = 0) => {
  const accounts = await fetchPredeployedAccounts()

  const account = accounts.at(accountIndex)
  if (account == null) {
    throw new Error('no accounts available')
  }

  return {
    address: account.address as string,
    privateKey: account.private_key as string,
    balance: parseInt(account.initial_balance, 10),
  }
}

export const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
) => {
  const deps: Dependencies = {
    logger: noopLogger,
    prompt: noopPrompt,
    makeEnv: async (flags) => {
      if (flags.pk == null || flags.account == null) {
        const acct = await fetchAccount()
        flags.account = flags.account ?? acct.address
        flags.pk = flags.pk ?? acct.privateKey
      }

      return {
        providerUrl: LOCAL_URL,
        pk: flags.pk as string,
        account: flags.account as string,
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
    makeEnv: async (flags) => {
      if (flags.pk == null || flags.account == null) {
        const acct = await fetchAccount()
        flags.account = flags.accout ?? acct.address
        flags.pk = flags.pk ?? acct.privateKey
      }

      return {
        providerUrl: LOCAL_URL,
        pk: flags.pk as string,
        account: flags.account as string,
      }
    },
    makeProvider: makeProvider,
  }
  return registerCommand(deps)
}
