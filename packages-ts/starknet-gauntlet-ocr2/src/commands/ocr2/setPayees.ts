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
  const rdd = getRDD(flags.rdd)
  const contractAddr = args[0]
  const aggregator = rdd[CONTRACT_TYPES.AGGREGATOR][contractAddr]
  if (!aggregator || !aggregator.payees)
    throw new Error(`Payees not defined in RDD for contract ${contractAddr}`)
  return aggregator.payees
}

const makeContractInput = async (
  input: UserInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
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
      `yarn gauntlet ocr2:set_payees --network=<NETWORK> --rdd=<RDD_PATH> <CONTRACT_ADDRESS>`,
    ],
  },
  validations: [],
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
