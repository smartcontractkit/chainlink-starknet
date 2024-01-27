import { Provider, CallContractResponse, constants } from 'starknet'

// Starknet network: Either goerli-alpha or mainnet-alpha
const network = 'goerli-alpha'

/**
 * Network: Starknet Goerli testnet
 * Aggregator: LINK/USD
 * Address: 0x2579940ca3c41e7119283ceb82cd851c906cbb1510908a913d434861fdcb245
 * Find more proxy address at:
 * https://docs.chain.link/data-feeds/price-feeds/addresses?network=starknet
 */
const dataFeedAddress = '0x2579940ca3c41e7119283ceb82cd851c906cbb1510908a913d434861fdcb245'

export async function readLatestRoundOffChain() {
  const provider = new Provider({
    sequencer: {
      network: constants.NetworkName.SN_GOERLI,
    },
  })

  const latestRound = await provider.callContract({
    contractAddress: dataFeedAddress,
    entrypoint: 'latest_round_data',
  })

  printResult(latestRound)
  return latestRound
}

function printResult(latestRound: CallContractResponse) {
  console.log('round_id =', parseInt(latestRound.result[0], 16))
  console.log('answer =', parseInt(latestRound.result[1], 16))
  console.log('block_num =', parseInt(latestRound.result[2], 16))
  console.log('observation_timestamp =', parseInt(latestRound.result[3], 16))
  console.log('transmission_timestamp =', parseInt(latestRound.result[4], 16))
}

readLatestRoundOffChain()
