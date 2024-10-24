import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  TOKEN = 'token',
}

export const tokenContractLoader = () => loadContract('chainlink_token_v1_link_token_LinkToken')
