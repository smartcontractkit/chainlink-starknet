import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { tokenContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
    return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
    contractId: CATEGORIES.TOKEN,
    category: CATEGORIES.TOKEN,
    action: 'declare',
    ux: {
        description: `Declares an ${CATEGORIES.TOKEN} contract`,
        examples: [`${CATEGORIES.TOKEN}:declare --network=<NETWORK>`],
    },
    makeUserInput,
    makeContractInput,
    validations: [],
    loadContract: tokenContractLoader,
}

export default makeExecuteCommand(commandConfig)
