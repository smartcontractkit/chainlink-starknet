import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { IStarknetProvider } from '../provider'
import { IWallet } from '../wallet'

interface Env {
  providerUrl: string
  pk?: string
}

export interface Dependencies {
  logger: typeof logger
  prompt: typeof prompt
  makeProvider: (url: string) => IStarknetProvider
  makeWallet: (pk?: string) => IWallet
  makeEnv: (flags: Record<string, string>) => Env
}
