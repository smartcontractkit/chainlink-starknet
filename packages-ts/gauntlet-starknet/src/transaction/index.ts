import { AddTransactionResponse } from 'starknet'

export type TransactionResponse = {
  hash: string
  address?: string
  wait: () => Promise<{ success: boolean }>
  tx?: AddTransactionResponse
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED'
  errorMessage?: string
}
