import {
  InvokeFunctionResponse,
  DeclareContractResponse,
  DeployContractResponse,
  TXN_STATUS,
} from 'starknet'

export type TransactionResponse = {
  hash: string
  address?: string
  wait: () => Promise<{ success: boolean }>
  tx?: InvokeFunctionResponse | DeclareContractResponse | DeployContractResponse
  code?: TXN_STATUS
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED'
  errorMessage?: string
}
