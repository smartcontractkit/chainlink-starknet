import {
  ExecuteCommandConfig,
  isValidAddress,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'

type ContractInput = [address: string]

export interface SetL1SenderInput {
  address: string
}

const validateAddress = async (input) => {
  if (!isValidAddress(input.address)) {
    throw new Error(`Invalid L1 Sender Address: ${input.address}`)
  }
  return true
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
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:set_l1_sender --network=<NETWORK> --address=<L1_SENDER_ADDRESS> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateAddress],
  loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
