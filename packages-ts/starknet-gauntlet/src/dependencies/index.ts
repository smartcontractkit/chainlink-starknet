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
  makeWallet: (
    withLedger: boolean,
    ledgerPath: string,
    pk?: string,
    account?: string,
  ) => Promise<IStarknetWallet>
  makeEnv: (flags: Record<string, string | boolean>) => Env
}

export type InspectionDependencies = Omit<Dependencies, 'makeWallet'>
