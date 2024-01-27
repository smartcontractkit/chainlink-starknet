import {
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ProxyLoader, CONTRACT_LIST } from '../../lib/contracts'
import { validateClassHash } from '../../lib/utils'

type UserInput = {
  owner: string
  address: string
  classHash?: string
}

type ContractInput = [owner: string, address: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
    address: flags.address,
    classHash: flags.classHash,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owner, input.address]
}

const validateAddresses = async (input) => {
  if (!isValidAddress(input.owner)) throw new Error(`Invalid owner address: ${input.owner}`)

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
    `About to deploy proxy for aggregator at ${input.user.address} with owner ${input.user.owner}`,
  )
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.PROXY,
  category: CATEGORIES.PROXY,
  action: 'deploy',
  ux: {
    description: 'Deploys an OCR2 aggregator proxy',
    examples: [
      `${CATEGORIES.PROXY}:deploy --network=<NETWORK> --address=<AGGREGATOR_ADDRESS> --classHash=<CLASS_HASH>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateAddresses, validateClassHash],
  loadContract: ocr2ProxyLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
