import { Call } from 'starknet'

export enum Action {
  APPROVE = 'approve',
  EXECUTE = 'execute',
  NONE = 'none',
}

export type State = {
  multisig: {
    address: string
    threshold: number
    signers: string[]
  }
  proposal?: {
    id: number
    nextAction: Action
    data: Call
    confirmations: number
  }
}
