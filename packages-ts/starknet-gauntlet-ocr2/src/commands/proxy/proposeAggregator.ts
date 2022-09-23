import {
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ProxyLoader, CONTRACT_LIST } from '../../lib/contracts'

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
  if (!isValidAddress(input.address))
    throw new Error(`Invalid aggregator address: ${input.address}`)

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(
    `About to propose aggregator ${input.user.address} for proxy at ${input.contract.join(', ')}`,
  )
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.PROXY,
  category: CATEGORIES.PROXY,
  action: 'propose_aggregator',
  ux: {
    description: 'Propose new aggregator for OCR2 proxy',
    examples: [
      `${CATEGORIES.PROXY}:propose_aggregator --network=<NETWORK> --address=<AGGREGATOR> <PROXY>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateAddresses],
  loadContract: ocr2ProxyLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
