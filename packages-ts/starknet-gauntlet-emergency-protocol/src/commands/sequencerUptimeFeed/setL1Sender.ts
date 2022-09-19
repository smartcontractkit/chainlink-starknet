import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'

type ContractInput = [address: string]

export interface SetL1SenderInput {
  address: string
}

const makeUserInput = async (flags): Promise<SetL1SenderInput> => {
  if (flags.input) return flags.input as SetL1SenderInput
  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: SetL1SenderInput): Promise<ContractInput> => {
  return [input.address]
}

const commandConfig: ExecuteCommandConfig<SetL1SenderInput, ContractInput> = {
  contractId: CONTRACT_LIST.SEQUENCER_UPTIME_FEED,
  category: CATEGORIES.SEQUENCER_UPTIME_FEED,
  action: 'set_l1_sender',
  ux: {
    description: 'Sets the L1 sender address on the SequencerUptimeFeed contract',
    examples: [
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:set_l1_sender --network=<NETWORK> --address=0x31982C9e5edd99bb923a948252167ea4BbC38AC1 0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
