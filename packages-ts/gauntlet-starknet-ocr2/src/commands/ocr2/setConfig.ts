import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, encoding, SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'
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

const makeContractInput = async (input: SetConfigInput): Promise<ContractInput> => {
  const oracles: Oracle[] = input.signers.map((o, i) => ({
    signer: input.signers[i],
    transmitter: input.transmitters[i],
  }))
  const { offchainConfig } = await encoding.serializeOffchainConfig(input.offchainConfig, input.secret)
  console.log('CONFIG', bytesToFelts(offchainConfig))
  return [oracles, new BN(input.f).toNumber(), input.onchainConfig, 2, bytesToFelts(offchainConfig)]
}

const commandConfig: ExecuteCommandConfig<SetConfigInput, ContractInput> = {
  ...SetConfig,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
