import {
  CONTRACT_TYPES,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  getRDD,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { CONTRACT_LIST, ocr2ContractLoader } from '../../lib/contracts'

type Payee = {
  transmitter: string
  payee: string
}

type PayeeConfig = {
  payees: Payee[]
}

type UserInput = Payee[]

type ContractInput = PayeeConfig

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  if (flags.rdd) {
    const rdd = await getRDD(flags.rdd)
    const contractAddress = args[0]
    const contract = rdd[CONTRACT_TYPES.AGGREGATOR][contractAddress]
    return contract.payees
  }

  const transmitters = flags.transmitters.split(',')
  const payees = flags.payees.split(',')
  if (transmitters.length != payees.length)
    throw new Error('Invalid input for payee config: number of transmitters and payees must match')

  return transmitters.map((transmitter, i) => ({
    transmitter,
    payee: payees[i],
  })) as UserInput
}

const makeContractInput = async (
  input: UserInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
  return {
    payees: input.map((payee: Payee) => ({
      transmitter: payee.transmitter,
      payee: payee.payee,
    })),
  } as ContractInput
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.OCR2,
  action: 'set_payees',
  category: CATEGORIES.OCR2,
  ux: {
    description: 'Set payees of OCR2 contract',
    examples: [
      `yarn gauntlet ocr2:set_payees --network=<NETWORK> --transmitters=[<ACCOUNTS>] --payees=[<ACCOUNTS>] <CONTRACT_ADDRESS>`,
      `yarn gauntlet ocr2:set_payees --network=<NETWORK> --rdd=<RDD_PATH> <CONTRACT_ADDRESS>`,
    ],
  },
  validations: [],
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
