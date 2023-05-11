import { loadContract } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  MULTISIG = 'Multisig',
}

export const contractLoader = () => loadContract(CONTRACT_LIST.MULTISIG)
