import {
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  address: string
}

type ContractInput = [address: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.address]
}

const validateAddresses = async (input) => {
  if (!isValidAddress(input.address)) throw new Error(`Invalid address: ${input.address}`)

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(
    `About to add access to ${input.user.address} for aggregator at ${input.contract.join(', ')}`,
  )
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.OCR2,
  category: CATEGORIES.OCR2,
  action: 'add_access',
  ux: {
    description: 'Add read access to aggregator',
    examples: [
      `${CATEGORIES.PROXY}:add_access --network=<NETWORK> --address=<ADDRESS> <AGGREGATOR>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateAddresses],
  loadContract: ocr2ContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
