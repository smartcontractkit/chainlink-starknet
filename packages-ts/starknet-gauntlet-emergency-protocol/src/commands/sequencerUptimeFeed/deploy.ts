import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'
import { toFelt } from 'starknet/dist/utils/number'
import { CATEGORIES } from '../../lib/categories'

type ContractInput = [initial_status: string, owner_address: string]

export interface UserInput {
  initialStatus: number
  owner?: string
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [toFelt(0), input.owner]
}

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
    initialStatus: flags.initialStatus,
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.SEQUENCER_UPTIME_FEED,
  category: CATEGORIES.SEQUENCER_UPTIME_FEED,
  action: 'deploy',
  ux: {
    description: 'Deploys a SequencerUptimeFeed contract',
    examples: [
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:deploy --initialStatus=false --network=<NETWORK>`,
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:deploy --initialStatus=false --ownerAddress=<STARKNET_ADDRESS> --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
