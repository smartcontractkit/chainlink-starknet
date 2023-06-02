import {
    ExecuteCommandConfig,
    isValidAddress,
    makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST, uptimeFeedContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
    return []
}


const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
    contractId: CONTRACT_LIST.SEQUENCER_UPTIME_FEED,
    category: CATEGORIES.SEQUENCER_UPTIME_FEED,
    action: 'declare',
    ux: {
        description: 'Declares a SequencerUptimeFeed contract',
        examples: [
            `${CATEGORIES.SEQUENCER_UPTIME_FEED}:declare --network=<NETWORK>`,
        ],
    },
    makeUserInput,
    makeContractInput,
    validations: [],
    loadContract: uptimeFeedContractLoader,
}

export default makeExecuteCommand(commandConfig)
