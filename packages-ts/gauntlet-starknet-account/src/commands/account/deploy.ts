import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
} from '@chainlink/gauntlet-starknet'
import { ec } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { accountContractLoader } from '../../lib/contracts'

type UserInput = {
  publicKey: string
  privateKey?: string
}

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  // If public key is not provided, generate a new address
  const keypair = ec.genKeyPair()
  const generatedPK = '0x' + keypair.getPrivate('hex')
  const pubkey = flags.publicKey || ec.getStarkKey(ec.getKeyPair(generatedPK))
  return {
    publicKey: pubkey,
    privateKey: !flags.publicKey && generatedPK,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  // If address is provided, deployContract should use that as adressSalt
  return [input.publicKey]
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to deploy an Account Contract with public key ${input.contract[0]}`)
  if (input.user.privateKey) {
    await deps.prompt(`The generated private key will be shown next, continue?`)
    deps.logger.line()

    deps.logger.info(`To sign future transactions, store the Private Key`)
    deps.logger.info(`PRIVATE_KEY: ${input.user.privateKey}`)

    deps.logger.line()
  }
}

const afterExecute: AfterExecute<UserInput, ContractInput> = (context, input, deps) => async (result) => {
  deps.logger.success(`Account contract located at ${result.responses[0].tx.address}`)
  return {
    publicKey: input.user.publicKey,
    privateKey: input.user.privateKey,
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.ACCOUNT,
    function: 'deploy',
    examples: [`${CATEGORIES.ACCOUNT}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  loadContract: accountContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
