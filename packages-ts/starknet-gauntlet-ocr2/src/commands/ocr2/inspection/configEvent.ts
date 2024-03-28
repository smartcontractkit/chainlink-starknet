import { IStarknetProvider } from '@chainlink/starknet-gauntlet'
import { hash } from 'starknet'

export const getLatestOCRConfigEvent = async (
  provider: IStarknetProvider,
  contractAddress: string,
) => {
  // get block number in which the latest config was set
  const res = await provider.provider.callContract({
    contractAddress,
    entrypoint: 'latest_config_details',
    calldata: [],
  })
  const latestConfigDetails = {
    configCount: parseInt(res[0]),
    blockNumber: parseInt(res[1]),
    configDigest: res[2],
  }
  // if no config has been set yet, return empty config
  if (latestConfigDetails.configCount === 0) return []

  const keyFilter = [hash.getSelectorFromName('ConfigSet')]
  const chunk = await provider.provider.getEvents({
    address: contractAddress,
    from_block: { block_number: latestConfigDetails.blockNumber },
    to_block: { block_number: latestConfigDetails.blockNumber },
    keys: [keyFilter],
    chunk_size: 10,
  })
  const events = chunk.events
  // if no config set events found in the given block, throw error
  // this should not happen if block number in latestConfigDetails is set correctly
  if (events.length === 0)
    throw new Error(`No ConfigSet events found in block number ${latestConfigDetails.blockNumber}`)

  // assume last event found is the latest config, in the event that multiple
  // set_config transactions ended up in the same block
  return events[events.length - 1].data
}
