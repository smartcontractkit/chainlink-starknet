import {
  EVMExecuteCommandConfig,
  EVMExecutionContext,
  makeEVMExecuteCommand,
} from '@chainlink/evm-gauntlet'
import { utils, constants } from 'ethers'
import { isValidAddress } from '@chainlink/evm-gauntlet'
import { CATEGORIES } from '../../../lib/categories'
import {
  l1BridgeContractLoader,
  l1BridgeProxyContractLoader,
  CONTRACT_LIST,
} from '../../../lib/contracts'

type UserInput = {
  l1BridgeAddress: string
  tokenAddress: string
  starknetMessagingAddress: string // starknet core contract implements the interface
}
type ContractInput = [l1BridgeAddress: string, encodedInput: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    l1BridgeAddress: flags.bridge,
    tokenAddress: flags.token,
    starknetMessagingAddress: flags.core,
  }
}

const makeContractInput = async (
  input: UserInput,
  context: EVMExecutionContext,
): Promise<ContractInput> => {
  const l1BridgeArtifact = await l1BridgeContractLoader()
  const l1BridgeInterface = l1BridgeArtifact.interface
  const data = utils.hexConcat([
    utils.hexZeroPad(constants.AddressZero, 32),
    utils.hexZeroPad(input.tokenAddress, 32),
    utils.hexZeroPad(input.starknetMessagingAddress, 32),
  ])
  const encodedInput = l1BridgeInterface.encodeFunctionData('initialize(bytes data)', [data])
  return [input.l1BridgeAddress, encodedInput]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!isValidAddress(input.l1BridgeAddress)) {
    throw new Error(`Invalid address of L1 bridge: ${input.l1BridgeAddress}`)
  }
  if (!isValidAddress(input.tokenAddress)) {
    throw new Error(`Invalid address of L1 token: ${input.tokenAddress}`)
  }
  if (!isValidAddress(input.starknetMessagingAddress)) {
    throw new Error(`Invalid address of Starknet core: ${input.starknetMessagingAddress}`)
  }
  return true
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE_PROXY,
  category: CATEGORIES.L1_BRIDGE,
  action: 'deploy',
  suffixes: ['proxy'],
  ux: {
    description: 'Deploys an L1 token bridge proxy and initialises the bridge',
    examples: [
      `${CATEGORIES.L1_BRIDGE}:deploy:proxy --network=<NETWORK> --bridge=<L1_BRIDGE_ADDRESS> --token=<L1_TOKEN_ADDRESS> --core=<STRAKNET_CORE_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l1BridgeProxyContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
