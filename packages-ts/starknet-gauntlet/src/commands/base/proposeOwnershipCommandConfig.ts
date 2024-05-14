import { isValidAddress } from '../../utils'
import { ExecuteCommandConfig } from '.'

type ContractInput = [newOwner: string]

export interface ProposeOwnerhipUserInput {
  newOwner: string
}

const validateNewOwner = async (input) => {
  if (isValidAddress(input.newOwner)) {
    return true
  }
  throw new Error(`Invalid New Owner Address: ${input.newOwner}`)
}

const makeUserInput = async (flags): Promise<ProposeOwnerhipUserInput> => {
  if (flags.input) return flags.input as ProposeOwnerhipUserInput
  return {
    newOwner: flags.newOwner,
  }
}

const makeContractInput = async (input: ProposeOwnerhipUserInput): Promise<ContractInput> => {
  return [input.newOwner]
}

export const proposeOwnershipCommandConfig = (
  contractId: string,
  category: string,
  contractLoader: any,
): ExecuteCommandConfig<ProposeOwnerhipUserInput, ContractInput> => ({
  contractId,
  category: contractId,
  // the on-chain method is called 'transfer_ownership' but naming the gauntlet command
  // 'propose_ownership' makes the intended purpose less ambiguous
  action: 'transfer_ownership',
  ux: {
    description: 'Begin two-step ownership transfer process by proposing pending owner',
    examples: [
      `${contractId}:propose_ownership --network=<NETWORK> --newOwner=<NEW_OWNER_ADDRESS> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateNewOwner],
  loadContract: contractLoader,
})
