import { BeforeExecute, isValidAddress } from '@chainlink/starknet-gauntlet'
import { uint256 } from 'starknet'
import { CATEGORIES } from './categories'
import { bridgeContractLoader, CONTRACT_LIST } from './contracts'

type UserInput = {
  account: string
}

type ContractInput = [account: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    account: flags.account,
  }
}

const validateAccount = async (input) => {
  if (!isValidAddress(input.account)) throw new Error(`Invalid account address: ${input.account}`)
  return true
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.account]
}

const beforeAdminRemoveExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(
    `About to remove admin ${input.user.account} for l2 bridge ${context.contractAddress}`,
  )
}

const beforeAdminRegisterExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(
    `About to register admin ${input.user.account} for l2 bridge ${context.contractAddress}`,
  )
}

export const createRegisterAdminCommandConfig = (action: string) => {
  return {
    contractId: CONTRACT_LIST.BRIDGE,
    category: CATEGORIES.BRIDGE,
    action,
    ux: {
      description: 'Registers admin',
      examples: [
        `${CATEGORIES.BRIDGE}:${action} --network=<NETWORK> --account=<ADMIN> <CONTRACT_ADDRESS>`,
      ],
    },
    makeUserInput,
    makeContractInput,
    validations: [validateAccount],
    loadContract: bridgeContractLoader,
    hooks: {
      beforeExecute: beforeAdminRegisterExecute,
    },
  }
}

export const createRemoveAdminCommandConfig = (action: string) => {
  return {
    contractId: CONTRACT_LIST.BRIDGE,
    category: CATEGORIES.BRIDGE,
    action,
    ux: {
      description: 'Removes admin',
      examples: [
        `${CATEGORIES.BRIDGE}:${action} --network=<NETWORK> --account=<ADMIN> <CONTRACT_ADDRESS>`,
      ],
    },
    makeUserInput,
    makeContractInput,
    validations: [validateAccount],
    loadContract: bridgeContractLoader,
    hooks: {
      beforeExecute: beforeAdminRemoveExecute,
    },
  }
}
