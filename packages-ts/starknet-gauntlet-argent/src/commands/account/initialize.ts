import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
} from '@chainlink/starknet-gauntlet'
import { ec } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { accountContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  publicKey: string
  privateKey?: string
}

type ContractInput = [string, 0]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  // If public key is not provided, generate a new address
  const keypair = ec.starkCurve.utils.randomPrivateKey()
  const generatedPK = '0x' + Buffer.from(keypair).toString('hex')
  const pubkey = flags.publicKey || ec.starkCurve.getStarkKey(keypair)
  return {
    publicKey: pubkey,
    privateKey: !flags.publicKey && generatedPK,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.publicKey, 0]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(
    `About to deploy an Argent Account Contract with public key ${input.contract[0]}`,
  )
  if (input.user.privateKey) {
    await deps.prompt(`The generated private key will be shown next, continue?`)
    deps.logger.line()

    deps.logger.info(`To sign future transactions, store the Private Key`)
    deps.logger.info(`PRIVATE_KEY: ${input.user.privateKey}`)

    deps.logger.line()
  }
}

const afterExecute: AfterExecute<UserInput, ContractInput> = (context, input, deps) => async (
  result,
) => {
  deps.logger.success(`Account contract located at ${result.responses[0].tx.address}`)
  return {
    publicKey: input.user.publicKey,
    privateKey: input.user.privateKey,
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'initialize',
  ux: {
    description:
      'Initializes an Argent Account Contract (Generates a public/private key for account)',
    examples: [
      `${CATEGORIES.ACCOUNT}:initialize --network=<NETWORK> --publicKey=<ADDRESS> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accountContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
