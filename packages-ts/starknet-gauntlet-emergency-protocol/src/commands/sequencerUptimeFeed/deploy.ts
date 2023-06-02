import {
  ExecuteCommandConfig,
  isValidAddress,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type ContractInput = [initial_status: number, owner_address: string]

export interface UserInput {
  classHash?: string
  initialStatus: number
  owner?: string
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.initialStatus, input.owner]
}

const validateOwner = async (input) => {
  if (!isValidAddress(input.owner)) {
    throw new Error(`Invalid Owner Address: ${input.owner}`)
  }
  return true
}

const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash) || input.classHash === undefined) {
    return true
  }
  throw new Error(`Invalid Owner Address: ${input.owner}`)
}

const validateInitialStatus = async (input) => {
  const status = Number(input.initialStatus)
  if (status !== 1 && status !== 0) {
    throw new Error(`Invalid Initial Status: ${input.initialStatus}`)
  }
  return true
}

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
    initialStatus: flags.initialStatus,
    classHash: flags.classHash
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.SEQUENCER_UPTIME_FEED,
  category: CATEGORIES.SEQUENCER_UPTIME_FEED,
  action: 'deploy',
  ux: {
    description: 'Deploys a SequencerUptimeFeed contract',
    examples: [
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:deploy --initialStatus=<INITIAL_STATUS> --network=<NETWORK>`,
      `${CATEGORIES.SEQUENCER_UPTIME_FEED}:deploy --initialStatus=<INITIAL_STATUS> --owner=<STARKNET_ADDRESS> --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateOwner, validateInitialStatus, validateClassHash],
  loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
