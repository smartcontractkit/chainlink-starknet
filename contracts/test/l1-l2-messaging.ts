import { ETH_DEVNET_URL, STARKNET_DEVNET_URL } from './constants'

//
// Docs: https://github.com/0xSpaceShard/starknet-devnet-rs/blob/main/contracts/l1-l2-messaging/README.md#ethereum-setup
//

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/models.rs#L23
 */
export type PostmanLoadL1MessagingContract = Readonly<{
  networkUrl?: string
  address?: string
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/models.rs#L132
 */
export type MessagingLoadAddress = Readonly<{
  messaging_contract_address: string
}>

/*
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/endpoints/postman.rs#L12
 */
export const loadL1MessagingContract = async (
  params?: PostmanLoadL1MessagingContract,
): Promise<MessagingLoadAddress> => {
  const res = await fetch(`${STARKNET_DEVNET_URL}/postman/load_l1_messaging_contract`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      network_url: params?.networkUrl ?? ETH_DEVNET_URL,
      address: params?.address,
    }),
  })

  const result = await res.json()
  if (result.error != null) {
    throw new Error(result.error)
  }
  return result
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
 * https://github.com/0xSpaceShard/starknet-devnet-rs/blob/7e5ff351198f799816c1857c1048bf8ee7f89428/crates/starknet-devnet-server/src/api/http/endpoints/postman.rs#L26
 */
export const flush = async (params?: FlushParameters): Promise<FlushedMessages> => {
  const res = await fetch(`${STARKNET_DEVNET_URL}/postman/flush`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      dry_run: params?.dryRun ?? false,
    }),
  })

  const result = await res.json()
  if (result.error != null) {
    throw new Error(result.error)
  }
  return result
}
