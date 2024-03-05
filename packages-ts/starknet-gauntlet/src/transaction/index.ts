import {
  InvokeFunctionResponse,
  DeclareContractResponse,
  DeployContractResponse,
  RPC,
} from 'starknet'

export type TransactionResponse = {
  hash: string
  address?: string
  wait: () => Promise<{ success: boolean }>
  tx?: InvokeFunctionResponse | DeclareContractResponse | DeployContractResponse
  code?: RPC.SPEC.TXN_STATUS
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED'
  errorMessage?: string
}
