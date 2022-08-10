import {
    BeforeExecute,
    ExecuteCommandConfig,
    makeExecuteCommand,
    isValidAddress,
  } from '@chainlink/starknet-gauntlet' // todo: use @chainlink/evm-gauntlet
  import { Uint256 } from 'starknet/dist/utils/uint256'
  import { bnToUint256 } from 'starknet/dist/utils/uint256'
  import { CATEGORIES } from '../../lib/categories'
  import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'
  
  type UserInput = {
    recipient: string,
    amount: string,
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
  
  const validateInput = async (input: UserInput): Promise<boolean> => {
    if (!isValidAddress(input.recipient)) {
      throw new Error(`Invalid address of L1 recipient: ${input.recipient}`)
    }
    return true
  }
  
  const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
    deps.logger.info(`About to withdraw from L1 bridge with the following details:
      ${input.contract}
    `)
  }
  
  const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
    contractId: CONTRACT_LIST.L1_BRIDGE,
    category: CATEGORIES.L1_BRIDGE,
    action: 'withdraw',
    ux: {
      description: 'Withdraws funds from L1 bridge to L1 recipient',
      examples: [
        `${CATEGORIES.L1_BRIDGE}:withdraw --network=<NETWORK> --recipient=[L1_RECIPIENT_ADDRESS] --amount=[AMOUNT] [L1_BRIDGE_ADDRESS]`,
      ],
    },
    makeUserInput,
    makeContractInput,
    validations: [validateInput],
    loadContract: l1BridgeContractLoader,
    hooks: {
      beforeExecute,
    },
  }
  
  export default makeExecuteCommand(commandConfig)
  