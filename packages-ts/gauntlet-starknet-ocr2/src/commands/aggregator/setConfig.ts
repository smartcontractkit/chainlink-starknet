import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader } from '../../lib/contracts'
import { ec, number } from 'starknet'

type OnchainConfig = any // TODO: Define more clearly
type OffchainConfig = any // TODO: Define more clearly

type Oracle = {
  signer: string
  transmitter: string
}

type UserInput = {
  f: number
  signers: string[]
  transmitters: string[]
  onchainConfig: OnchainConfig
  offchainConfig: OffchainConfig
  offchainConfigVersion: number
}

type ContractInput = [
  {
    oracles: Oracle[]
    f: number
    onchain_config: string
    offchain_config_version: number
    offchain_config: string
  },
]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    f: flags.f,
    signers: flags.signers,
    transmitters: flags.transmitters,
    onchainConfig: flags.onchainConfig || 1,
    offchainConfig: flags.offchainConfig || [1],
    offchainConfigVersion: flags.offchainConfigVersion || 2,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  const oracles: Oracle[] = input.signers.map((o, i) => ({
    signer: number.toBN(ec.getStarkKey(o)),
    transmitter: input.transmitters[i],
  }))
  return [
    {
      oracles: oracles,
      f: new BN(input.f).toNumber(),
      onchain_config: input.onchainConfig.toString('base64'),
      offchain_config_version: 2,
      offchain_config: input.offchainConfig.toString('base64'),
    },
  ]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (3 * input.f >= input.signers.length)
    throw new Error(`Signers length needs to be higher than 3 * f (${3 * input.f}). Currently ${input.signers.length}`)

  if (input.signers.length !== input.transmitters.length)
    throw new Error(`Signers and Trasmitters length are different`)

  // TODO: Add validations for offchain config
  return true
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'set_config',
    examples: [`${CATEGORIES.OCR2}:set_config --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
