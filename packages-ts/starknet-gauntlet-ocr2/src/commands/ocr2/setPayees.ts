import {
  CONTRACT_TYPES,
  ExecuteCommandConfig,
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
    // Read inputs
    const contractAddress = args[0]
    const rdd = getRDD(flags.rdd)

    // Get the "operators" section of the RDD file
    const operators = rdd.operators
    if (operators == null || typeof operators !== 'object') {
      throw new Error(`expected rdd["operators"] to be an object: ${operators}`)
    }

    // Get the config that corresponds to the input contract address
    const contract = rdd?.[CONTRACT_TYPES.AGGREGATOR]?.[contractAddress]
    if (contract == null || typeof contract !== 'object') {
      throw new Error(
        `expected rdd["${CONTRACT_TYPES.AGGREGATOR}"]["${contractAddress}"] to be an object: ${contract}`,
      )
    }

    // Get the contract's oracles
    const oracles = contract.oracles
    if (oracles == null || !Array.isArray(oracles)) {
      throw new Error(
        `expected rdd["${CONTRACT_TYPES.AGGREGATOR}"]["${contractAddress}"]["oracles"] to be an array: ${oracles}`,
      )
    }

    // Iterate over the contract's oracles
    return oracles.map((oracle, i) => {
      // Get the operator name from the oracle
      const operatorName = oracle.operator
      if (operatorName == null || typeof operatorName !== 'string') {
        throw new Error(
          `expected rdd["${CONTRACT_TYPES.AGGREGATOR}"]["${contractAddress}"]["oracles"][${i}]["operator"] to be a string: ${operatorName}`,
        )
      }

      // Use the operator name to get the transmitter and payee info from the "operators" section of the RDD file
      const operator = operators[operatorName]
      if (operator == null || typeof operator !== 'object') {
        throw new Error(`expected rdd["operators"]["${operatorName}"] to be an object: ${operator}`)
      }

      // Return the transmitter and payee info
      return {
        transmitter: operator.ocrNodeAddress?.[0],
        payee: operator.payeeAddress,
      } as Payee
    })
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

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
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
