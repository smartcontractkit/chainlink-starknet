import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  TOKEN = 'token',
}

export const tokenContractLoader = () => loadContract('LinkToken')
