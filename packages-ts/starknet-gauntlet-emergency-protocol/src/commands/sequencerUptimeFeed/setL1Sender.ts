import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'

type ContractInput = [string]

export interface SetL1SenderInput {
  l1Sender: string
}

const makeUserInput = async (flags): Promise<SetL1SenderInput> => {
  if (flags.input) return flags.input as SetL1SenderInput
  return {
    l1Sender: flags.l1Sender,
  }
}

const makeContractInput = async (input: SetL1SenderInput): Promise<ContractInput> => {
  return [input.l1Sender]
}

const commandConfig: ExecuteCommandConfig<SetL1SenderInput, ContractInput> = {
  contractId: CONTRACT_LIST.SEQUENCER_UPTIME_FEED,
  category: CATEGORIES.SEQUENCER_UPTIME_FEED,
  action: 'set_l1_sender',
  ux: {
    description: 'Sets the L1 sender address on the SequencerUptimeFeed contract',
    examples: [
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:set_l1_sender --l1Sender=0x31982C9e5edd99bb923a948252167ea4BbC38AC1 --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
