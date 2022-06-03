import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'

type Oracle = {
  signer: string
  transmitter: string
}

type ContractInput = [
  // oracles_len: any,
  oracles: Oracle[],
  f: number,
  onchain_config: string,
  offchain_config_version: number,
  offchain_config: string,
]

const makeUserInput = async (flags, args): Promise<SetConfigInput> => {
  if (flags.input) return flags.input as SetConfigInput

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

const makeContractInput = async (input: SetConfigInput): Promise<ContractInput> => {
  const oracles: Oracle[] = input.signers.map((o, i) => ({
    signer: input.signers[i],
    transmitter: input.transmitters[i],
  }))
  return [oracles, new BN(input.f).toNumber(), input.onchainConfig, 2, input.offchainConfig]
}

const commandConfig: ExecuteCommandConfig<SetConfigInput, ContractInput> = {
  ...SetConfig,
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
