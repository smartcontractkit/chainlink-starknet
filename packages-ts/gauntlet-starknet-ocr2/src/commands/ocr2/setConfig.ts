import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/starknet-gauntlet'
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
  // oracles_len: any,
  oracles: Oracle[],
  f: number,
  onchain_config: string,
  offchain_config_version: number,
  offchain_config: string,
]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  if (flags.default) {
    // TODO: Remove this at some point and replace with some
    let f = 1
    let onchainConfig = 1
    let offchainConfigVersion = 2
    let offchainConfig = [93, 11111, 22222, 33333]

    return {
      f: f,
      signers: [
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733',
      ],
      transmitters: [
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732',
        '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733',
      ],
      onchainConfig: onchainConfig,
      offchainConfig: offchainConfig,
      offchainConfigVersion: offchainConfigVersion,
    }
  }

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
    signer: input.signers[i],
    transmitter: input.transmitters[i],
  }))
  return [oracles, new BN(input.f).toNumber(), input.onchainConfig, 2, input.offchainConfig]
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
    examples: [
      `${CATEGORIES.OCR2}:set_config --network=<NETWORK> --address=<ADDRESS> --f=<NUMBER> --signers=[<ACCOUNTS>] --transmitters=[<ACCOUNTS>] --onchainConfig=<CONFIG> --offchainConfig=<CONFIG> --offchainConfigVersion=<NUMBER> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
