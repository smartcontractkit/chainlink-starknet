import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { IStarknetProvider } from '../provider'
import { IStarknetWallet } from '../wallet'

export interface Env {
  providerUrl: string
  pk?: string
  account?: string
}

export interface Dependencies {
  logger: typeof logger
  prompt: typeof prompt
  makeProvider: (url: string) => IStarknetProvider
  makeWallet: (pk: string) => IStarknetWallet
  makeEnv: (flags: Record<string, string>) => Env
}
