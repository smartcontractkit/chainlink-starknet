import { api } from 'starknet'

export type TransactionResponse = {
  hash: string
  address?: string
  wait: () => Promise<{ success: boolean }>
  tx?: api.Sequencer.AddTransactionResponse
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED'
  errorMessage?: string
}
