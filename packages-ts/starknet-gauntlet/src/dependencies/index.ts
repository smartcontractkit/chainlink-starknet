import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { IStarknetProvider } from '../provider'
import { IStarknetWallet } from '../wallet'

export interface Env {
  providerUrl: string
  pk?: string
  account?: string
  withLedger?: boolean
  ledgerPath?: string
  multisig?: string
  [key: string]: string | boolean // Custom env
}

export interface Dependencies {
  logger: typeof logger
  prompt: typeof prompt
  makeProvider: (url: string) => IStarknetProvider
  makeEnv: (flags: Record<string, string | boolean>) => Env
  makeWallet: (env: Env) => Promise<IStarknetWallet>
}

export type InspectionDependencies = Omit<Dependencies, 'makeWallet'>
