import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { aggregatorProxyLoader, CONTRACT_LIST } from '../../lib/contracts'

export interface UserInput {
  owner: string
  aggregator: string
}

type ContractInput = [owner: string, aggregator: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
    aggregator: flags.aggregator,
  } as UserInput
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owner, input.aggregator]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.AGGREGATOR_PROXY,
  category: CATEGORIES.AGGREGATOR_PROXY,
  action: 'deploy',
  ux: {
    description: 'Deploys an Aggregator Proxy contract',
    examples: [
      `${CATEGORIES.AGGREGATOR_PROXY}:deploy --aggregator=0x48a792b964dd871ed09d6d0f980b8d0f50d48b54d4bd14ab595a7486be1cbdf --network=<NETWORK>`,
    ],
  },
  validations: [],
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: aggregatorProxyLoader,
}

export default makeExecuteCommand(commandConfig)
