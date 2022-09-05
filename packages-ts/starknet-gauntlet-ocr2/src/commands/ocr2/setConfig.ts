import {
  AfterExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
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
  onchain_config: string[],
  offchain_config_version: number,
  offchain_config: string[],
]

const makeContractInput = async (input: SetConfigInput): Promise<ContractInput> => {
  const oracles: Oracle[] = input.signers.map((o, i) => {
    // standard format from chainlink node ocr2on_starknet_<key> (no 0x prefix)
    let signer = input.signers[i].replace('ocr2on_starknet_', '') // replace prefix if present
    signer = signer.startsWith('0x') ? signer : '0x' + signer // prepend 0x if missing

    return {
      signer,
      transmitter: input.transmitters[i],
    }
  })

  // remove prefix if present on offchain key
  input.offchainConfig.offchainPublicKeys = input.offchainConfig.offchainPublicKeys.map((k) =>
    k.replace('ocr2off_starknet_', ''),
  )
  input.offchainConfig.configPublicKeys = input.offchainConfig.configPublicKeys.map((k) =>
    k.replace('ocr2cfg_starknet_', ''),
  )

  const { offchainConfig } = await encoding.serializeOffchainConfig(
    input.offchainConfig,
    input.secret,
  )
  let onchainConfig = [] // onchain config should be empty array for input (generate onchain)
  return [oracles, new BN(input.f).toNumber(), onchainConfig, 2, bytesToFelts(offchainConfig)]
}

const afterExecute: AfterExecute<SetConfigInput, ContractInput> = (context, input, deps) => async (
  result,
) => {
  const txHash = result.responses[0].tx.hash
  const txInfo = await context.provider.provider.getTransactionReceipt(txHash)
  const eventData = (txInfo.events[0] as any).data

  const offchainConfig = decodeOffchainConfigFromEventData(eventData)
  try {
    // remove cfg keys from user input
    delete input.user.offchainConfig.configPublicKeys
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
