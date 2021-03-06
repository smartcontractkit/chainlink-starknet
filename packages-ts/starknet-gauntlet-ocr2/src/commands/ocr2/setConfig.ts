import { AfterExecute, ExecuteCommandConfig, makeExecuteCommand, BeforeExecute } from '@chainlink/starknet-gauntlet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, encoding, SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'
import { bytesToFelts, decodeOffchainConfigFromEventData } from '../../lib/encoding'
import assert from 'assert'

type Oracle = {
  signer: string
  transmitter: string
}

type ContractInput = [
  oracles: Oracle[],
  f: number,
  onchain_config: string,
  offchain_config_version: number,
  offchain_config: string[],
]

const makeContractInput = async (input: SetConfigInput): Promise<ContractInput> => {
  const oracles: Oracle[] = input.signers.map((o, i) => ({
    signer: input.signers[i],
    transmitter: input.transmitters[i],
  }))
  const { offchainConfig } = await encoding.serializeOffchainConfig(input.offchainConfig, input.secret)
  return [oracles, new BN(input.f).toNumber(), input.onchainConfig, 2, bytesToFelts(offchainConfig)]
}

const afterExecute: AfterExecute<SetConfigInput, ContractInput> = (context, input, deps) => async (result) => {
  const txHash = result.responses[0].tx.hash
  const txInfo = await context.provider.provider.getTransactionReceipt(txHash)
  const eventData = (txInfo.events[0] as any).data
  const offchainConfig = decodeOffchainConfigFromEventData(eventData)
  try {
    assert.deepStrictEqual(offchainConfig, input.user.offchainConfig)
    deps.logger.success('Configuration was successfully set')
    return { successfulConfiguration: true }
  } catch (e) {
    deps.logger.error('Configuration set is different than provided')
    deps.logger.log(offchainConfig)
    return { successfulConfiguration: false }
  }
}

const commandConfig: ExecuteCommandConfig<SetConfigInput, ContractInput> = {
  ...SetConfig,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
  hooks: {
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
