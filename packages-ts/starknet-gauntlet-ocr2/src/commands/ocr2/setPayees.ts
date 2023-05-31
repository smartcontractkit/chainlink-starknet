import {
  CONTRACT_TYPES,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  getRDD,
} from '@chainlink/starknet-gauntlet'
import { ocr2ContractLoader } from '../../lib/contracts'

type PayeeConfig = {
  transmitter: string
  payee: string
}

type UserInput = PayeeConfig[]

type ContractInput = [payees: PayeeConfig[]]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  if (args.length % 2 != 0) throw new Error('Invalid number of arguments for payee config')
  let userInput: UserInput = []
  for (let i = 0; i < args.length; i += 2) {
    userInput.push({
      transmitter: args[i],
      payee: args[i + 1],
    })
  }
  return userInput
}

const makeContractInput = async (
  input: UserInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
  if (ctx.rdd) {
    const aggregator = ctx.rdd[CONTRACT_TYPES.AGGREGATOR][ctx.contractAddress]
    input = aggregator.payees
  }

  return input.map((payeeConfig) => [
    {
      transmitter: payeeConfig.transmitter,
      payee: payeeConfig.payee,
    },
  ]) as ContractInput
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: 'ocr2',
  action: 'set_payees',
  category: 'ocr2',
  ux: {
    description: 'Set payees of OCR2 contract',
    examples: [
      `yarn gauntlet ocr2:set_payees --network=<NETWORK> --rdd=<RDD_PATH> <CONTRACT_ADDRESS> <TRANSMITTER_1> <PAYEE_1> <TRANSMITTER_2> <PAYEE_2> ...`,
    ],
  },
  validations: [],
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
