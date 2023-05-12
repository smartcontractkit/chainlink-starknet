import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  EXAMPLE = 'example',
}

export const tokenContractLoader = () => loadContract(CONTRACT_LIST.EXAMPLE)
