import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  SEQUENCER_UPTIME_FEED = 'sequencer_uptime_feed',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/../../contract_artifacts/abi/${name}.json`).toString('ascii'),
  )
}

export const uptimeFeedContractLoader = () => loadContract(CONTRACT_LIST.SEQUENCER_UPTIME_FEED)
