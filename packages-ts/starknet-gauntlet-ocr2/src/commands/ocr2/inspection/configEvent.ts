import { IStarknetProvider } from '@chainlink/starknet-gauntlet'
import { hash } from 'starknet'

export const getLatestOCRConfigEvent = async (
  provider: IStarknetProvider,
  contractAddress: string,
) => {
  // get block number in which the latest config was set
  const res = (
    await provider.provider.callContract({
      contractAddress,
      entrypoint: 'latest_config_details',
      calldata: [],
    })
  ).result
  const latestConfigDetails = {
    configCount: parseInt(res[0]),
    blockNumber: parseInt(res[1]),
    configDigest: res[2],
  }
  // if no config has been set yet, return empty config
  if (latestConfigDetails.configCount === 0) return []

  // retrieve block traces
  const blockTraces = await provider.provider.getBlockTraces(latestConfigDetails.blockNumber)
  // get list of all events across all internal calls in the block for which the key matches 'ConfigSet'
  const configSetEvents = blockTraces.traces
    .map((trace) => {
      return trace.function_invocation.internal_calls
        .filter((call) => call.contract_address === contractAddress)
        .map((call) => call.events)
        .flat()
        .filter((event) => event.keys[0] === hash.getSelectorFromName('ConfigSet'))
    })
    .flat()

  // assume last event found is the latest config set event
  return configSetEvents[configSetEvents.length - 1].data
}
