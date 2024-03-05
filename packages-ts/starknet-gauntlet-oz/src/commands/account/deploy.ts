import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { ec } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { accountContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  publicKey: string
  privateKey?: string
  salt?: number
  classHash?: string
}

type ContractInput = [publicKey: string]

const makeUserInput = async (flags, _, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  // If public key is not provided, generate a new address
  const keypair = ec.starkCurve.utils.randomPrivateKey()
  const generatedPK = '0x' + Buffer.from(keypair).toString('hex')
  const pubkey = flags.publicKey || env.publicKey || ec.starkCurve.getStarkKey(keypair)
  const salt: number = flags.salt ? +flags.salt : undefined
  return {
    publicKey: pubkey,
    privateKey: (!flags.publicKey || !env.account) && generatedPK,
    salt,
    classHash: flags.classHash,
  }
}

const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash) || input.classHash === undefined) {
    return true
  }
  throw new Error(`Invalid Class Hash: ${input.classHash}`)
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.publicKey]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy an OZ 0.x Account Contract with:
    public key: ${input.contract[0]}
    salt: ${input.user.salt || 'randomly generated'}
    action: ${context.action}`)
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

  return {
    publicKey: input.user.publicKey,
    privateKey: input.user.privateKey,
  }
}

const deployCommandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'deploy',
  ux: {
    description: 'Deploys an OpenZeppelin Account contract from an existing account',
    examples: [
      `${CATEGORIES.ACCOUNT}:deploy --network=<NETWORK> --address=<ADDRESS> --classHash=<CLASS_HASH> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateClassHash],
  loadContract: accountContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

const deployAccountCommandConfig: ExecuteCommandConfig<UserInput, ContractInput> = Object.assign(
  {},
  deployCommandConfig,
  {
    action: 'deploy-account',
    ux: {
      description: 'Deploys an OpenZeppelin Account contract using DEPLOY_ACCOUNT',
      examples: [
        `${CATEGORIES.ACCOUNT}:deploy-account --network=<NETWORK> --address=<ADDRESS> --classHash=<CLASS_HASH> <CONTRACT_ADDRESS>`,
      ],
    },
  },
)

const Deploy = makeExecuteCommand(deployCommandConfig)
const DeployAccount = makeExecuteCommand(deployAccountCommandConfig)

export { Deploy, DeployAccount }
