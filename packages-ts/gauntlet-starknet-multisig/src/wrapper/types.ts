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
    owners: string[]
  }
  proposal?: {
    id: number
    nextAction: Action
    data: Call
    approvers: number
  }
}
