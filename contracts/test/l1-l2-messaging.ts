import { ETH_DEVNET_URL, STARKNET_DEVNET_URL } from './constants'

//
// Docs: https://0xspaceshard.github.io/starknet-devnet/docs/postman
//

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/main/crates/starknet-devnet-server/src/api/http/models.rs#L23
 */
export type PostmanLoadL1MessagingContract = Readonly<{
  networkUrl?: string
  address?: string
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/main/crates/starknet-devnet-server/src/api/http/models.rs#L132
 */
export type MessagingLoadAddress = Readonly<{
  messaging_contract_address: string
}>

type JsonRpcResponse<T> = {
  result?: T
  error?: { message: string }
}

const devnetRpc = async <T>(method: string, params: Record<string, unknown>): Promise<T> => {
  const res = await fetch(`${STARKNET_DEVNET_URL}/rpc`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method,
      params,
    }),
  })

  const data = (await res.json()) as JsonRpcResponse<T>
  if (data.error != null) {
    throw new Error(data.error.message)
  }
  if (data.result == null) {
    throw new Error(`${method} returned no result`)
  }
  return data.result
}

/*
 * starknet-devnet-rs 0.8+ removed HTTP /postman/load_l1_messaging_contract; use devnet_postmanLoad.
 */
export const loadL1MessagingContract = async (
  params?: PostmanLoadL1MessagingContract,
): Promise<MessagingLoadAddress> => {
  return devnetRpc<MessagingLoadAddress>('devnet_postmanLoad', {
    network_url: params?.networkUrl ?? ETH_DEVNET_URL,
    ...(params?.address != null && { messaging_contract_address: params.address }),
  })
}

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/models.rs#L127
 */
export type FlushParameters = Readonly<{
  dryRun?: boolean
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-types/src/rpc/messaging.rs#L52
 */
export type MessageToL1 = Readonly<{
  from_address: string
  to_address: string
  payload: string[]
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-types/src/rpc/messaging.rs#L14
 */
export type MessageToL2 = Readonly<{
  l2_contract_address: string
  entry_point_selector: string
  l1_contract_address: string
  payload: string
  paid_fee_on_l1: string
  nonce: string
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/models.rs#L120
 */
export type FlushedMessages = Readonly<{
  messages_to_l1: MessageToL1[]
  messages_to_l2: MessageToL2[]
  generated_l2_transactions: string[]
  l1_provider: string
}>

/*
 * starknet-devnet-rs 0.8+ removed HTTP /postman/flush; use devnet_postmanFlush.
 */
export const flush = async (params?: FlushParameters): Promise<FlushedMessages> => {
  return devnetRpc<FlushedMessages>('devnet_postmanFlush', {
    dry_run: params?.dryRun ?? false,
  })
}
