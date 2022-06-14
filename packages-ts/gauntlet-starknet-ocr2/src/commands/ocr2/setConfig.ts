import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, encoding } from '@chainlink/gauntlet-contracts-ocr2'
import { bytesToFelts } from '../../lib/encoding'

type Oracle = {
  signer: string
  transmitter: string
}

type ContractInput = [
  oracles: Oracle[],
  f: number,
  onchain_config: string,
  offchain_config_version: number,
  offchain_config: BN[],
]

export interface SetConfigInput {
  f: number
  signers: string[]
  transmitters: string[]
  onchainConfig: string
  offchainConfig: encoding.OffchainConfig
  offchainConfigVersion: number
  secret: string
  randomSecret?: string
}

const makeUserInput = async (flags, args, env): Promise<SetConfigInput> => {
  if (flags.input) return flags.input as SetConfigInput

  if (flags.default) {
    // TODO: Remove this at some point and replace with some
    let f = 1
    let offchainConfigVersion = 2
    let offchainConfig = {} as encoding.OffchainConfig

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
      onchainConfig: '',
      offchainConfig: offchainConfig,
      offchainConfigVersion: offchainConfigVersion,
      secret: env.secret || flags.secret,
    }
  }

  return {
    f: flags.f,
    signers: flags.signers,
    transmitters: flags.transmitters,
    onchainConfig: flags.onchainConfig || 1,
    offchainConfig: flags.offchainConfig,
    offchainConfigVersion: flags.offchainConfigVersion || 2,
    secret: env.secret || flags.secret,
    randomSecret: flags.randomSecret,
  }
}

const makeContractInput = async (input: SetConfigInput): Promise<ContractInput> => {
  console.log(input)

  const oracles: Oracle[] = input.signers.map((o, i) => ({
    signer: input.signers[i],
    transmitter: input.transmitters[i],
  }))
  const { offchainConfig } = await encoding.serializeOffchainConfig(input.offchainConfig, input.secret)
  return [oracles, new BN(input.f).toNumber(), input.onchainConfig, 2, bytesToFelts(offchainConfig)]
}

const commandConfig: ExecuteCommandConfig<SetConfigInput, ContractInput> = {
  ...SetConfig,
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
