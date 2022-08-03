import { BeforeExecute, ExecuteCommandConfig, ExecutionContext, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'
import { validateSigners } from './deploy'

type UserInput = {
  signers: string[]
}

type ContractInput = [signers: string[]]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    signers: flags.signers,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.signers.map((addr) => number.toFelt(number.toBN(addr)))]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger
    .info(`About to set signers on the Multisig Contract (${context.contractAddress}) with the following details:
    - New Signers: ${input.user.signers}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    description: 'Set signers of the multisig account',
    examples: [`${CATEGORIES.MULTISIG}:set_signers --network=<NETWORK> --signers=[SIGNERS_LIST]`],
  },
  category: CATEGORIES.MULTISIG,
  action: 'set_signers',
  contractId: CATEGORIES.MULTISIG,
  makeUserInput,
  makeContractInput,
  validations: [validateSigners],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
