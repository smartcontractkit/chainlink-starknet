import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  TOKEN = 'LinkToken',
}

export const tokenContractLoader = () => loadContract(CONTRACT_LIST.TOKEN)
