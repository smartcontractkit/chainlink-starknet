import { Sequencer } from 'starknet'

export type TransactionResponse = {
  hash: string
  address?: string
  wait: () => Promise<{ success: boolean }>
  tx?: Sequencer.AddTransactionResponse
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED'
  errorMessage?: string
}
