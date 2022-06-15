import { BeforeExecute, ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/starknet-gauntlet'
import { Uint256 } from 'starknet/dist/utils/uint256'
import { bnToUint256 } from 'starknet/dist/utils/uint256'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'

type UserInput = {
  recipient: string
  amount: string
}

type ContractInput = [recipient: string, amount: Uint256]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    recipient: flags.recipient,
    amount: flags.amount,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.recipient, bnToUint256(input.amount)]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to mint an ERC20 Token Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.TOKEN,
    function: 'mint',
    examples: [
      `${CATEGORIES.TOKEN}:mint --network=<NETWORK> --link`,
      `${CATEGORIES.TOKEN}:mint --network=<NETWORK> --link`,
    ],
  },
  internalFunction: 'permissionedMint',
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
