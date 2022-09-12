import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { ec, hash } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import {
  accountContractLoader,
  CONTRACT_LIST,
  calculateAddress,
  equalAddress,
} from '../../lib/contracts'

type UserInput = {
  publicKey: string
  privateKey?: string
  salt?: number
}

type ContractInput = [publicKey: string]

const makeUserInput = async (flags, _, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  // If public key is not provided, generate a new address
  const keypair = ec.genKeyPair()
  const generatedPK = '0x' + keypair.getPrivate('hex')
  const pubkey = flags.publicKey || env.publicKey || ec.getStarkKey(ec.getKeyPair(generatedPK))
  const salt: number = flags.salt ? +flags.salt : undefined
  return {
    publicKey: pubkey,
    privateKey: (!flags.publicKey || !env.account) && generatedPK,
    salt,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.publicKey]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy an Account Contract with:
    public key: ${input.contract[0]}
    salt: ${input.user.salt || 'randomly generated'}`)
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
  const contract = result.responses[0].tx.address
  contract
    ? deps.logger.success(`Account contract located at ${contract}`)
    : deps.logger.error('Account contract deployment failed')

  if (input.user.salt != undefined) {
    const calcAddr = calculateAddress(input.user.salt, input.user.publicKey)

    // log error if address mismatch
    if (!equalAddress(contract, calcAddr)) {
      deps.logger
        .error(`Deployed account contract address (${contract}) does not match calculated account contract address (${calcAddr}).
        There is likely a difference in hash of the deployed contract and the CONTRACT_HASH constant`)
      deps.logger.warn(
        `Account addresses must match otherwise this could cause mismatched keys with chainlink node.`,
      )
    } else {
      deps.logger.success(`Deployed account matches expected contract address`)
    }
  }

  return {
    publicKey: input.user.publicKey,
    privateKey: input.user.privateKey,
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'deploy',
  ux: {
    description: 'Deploys an OpenZeppelin Account contract',
    examples: [
      `${CATEGORIES.ACCOUNT}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`,
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
