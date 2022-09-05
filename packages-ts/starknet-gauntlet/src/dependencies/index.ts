import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { IStarknetProvider } from '../provider'
import { IStarknetWallet } from '../wallet'
import { makeProvider } from '@chainlink/evm-gauntlet'

export interface Env {
  providerUrl: string
  pk?: string
  account?: string
  [key: string]: string // Custom env
}

export interface Dependencies {
  logger: typeof logger
  prompt: typeof prompt
  makeProvider: (url: string) => IStarknetProvider
  makeWallet: (pk: string, account?: string) => IStarknetWallet
  makeEnv: (flags: Record<string, string>) => Env
}

export type InspectionDependencies = Omit<Dependencies, 'makeWallet'>
