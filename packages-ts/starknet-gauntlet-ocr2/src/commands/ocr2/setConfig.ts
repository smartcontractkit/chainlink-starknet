import {
  CONTRACT_TYPES,
  AfterExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { time, BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, encoding, SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'
import { bytesToFelts, getRDD } from '@chainlink/starknet-gauntlet'
import { decodeOffchainConfigFromEventData } from '../../lib/encoding'
import assert from 'assert'
import { InvokeTransactionReceiptResponse } from 'starknet'

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

const makeUserInput = async (flags, args, env): Promise<SetConfigInput> => {
  if (flags.input) return flags.input as SetConfigInput
  if (flags.rdd) {
    const rdd = getRDD(flags.rdd)
    const contractAddr = args[0]
    const contract = rdd[CONTRACT_TYPES.AGGREGATOR][contractAddr]
    const config = contract.config
    const operators = contract.oracles.map((oracle) => rdd.operators[oracle.operator])
    const publicKeys = operators.map((o) =>
      o.ocr2OffchainPublicKey[0].replace('ocr2off_starknet_', ''),
    )
    const peerIds = operators.map((o) => o.peerId[0])
    const operatorConfigPublicKeys = operators.map((o) =>
      o.ocr2ConfigPublicKey[0].replace('ocr2cfg_starknet_', ''),
    )
    const offchainConfig: encoding.OffchainConfig = {
      deltaProgressNanoseconds: time.durationToNanoseconds(config.deltaProgress).toNumber(),
      deltaResendNanoseconds: time.durationToNanoseconds(config.deltaResend).toNumber(),
      deltaRoundNanoseconds: time.durationToNanoseconds(config.deltaRound).toNumber(),
      deltaGraceNanoseconds: time.durationToNanoseconds(config.deltaGrace).toNumber(),
      deltaStageNanoseconds: time.durationToNanoseconds(config.deltaStage).toNumber(),
      rMax: config.rMax,
      s: config.s,
      offchainPublicKeys: publicKeys,
      peerIds: peerIds,
      reportingPluginConfig: {
        alphaReportInfinite: config.reportingPluginConfig.alphaReportInfinite,
        alphaReportPpb: Number(config.reportingPluginConfig.alphaReportPpb),
        alphaAcceptInfinite: config.reportingPluginConfig.alphaAcceptInfinite,
        alphaAcceptPpb: Number(config.reportingPluginConfig.alphaAcceptPpb),
        deltaCNanoseconds: time
          .durationToNanoseconds(config.reportingPluginConfig.deltaC)
          .toNumber(),
      },
      maxDurationQueryNanoseconds: time.durationToNanoseconds(config.maxDurationQuery).toNumber(),
      maxDurationObservationNanoseconds: time
        .durationToNanoseconds(config.maxDurationObservation)
        .toNumber(),
      maxDurationReportNanoseconds: time.durationToNanoseconds(config.maxDurationReport).toNumber(),
      maxDurationShouldAcceptFinalizedReportNanoseconds: time
        .durationToNanoseconds(config.maxDurationShouldAcceptFinalizedReport)
        .toNumber(),
      maxDurationShouldTransmitAcceptedReportNanoseconds: time
        .durationToNanoseconds(config.maxDurationShouldTransmitAcceptedReport)
        .toNumber(),
      configPublicKeys: operatorConfigPublicKeys,
    }

    return {
      f: config.f,
      signers: operators.map((o) => o.ocr2OnchainPublicKey[0]),
      transmitters: operators.map((o) => o.ocrNodeAddress[0]),
      onchainConfig: [],
      offchainConfig,
      offchainConfigVersion: 2,
      secret: flags.secret || env.secret,
      randomSecret: flags.randomSecret || undefined,
    }
  }

  return {
    f: flags.f,
    signers: flags.signers,
    transmitters: flags.transmitters,
    onchainConfig: flags.onchainConfig,
    offchainConfig: flags.offchainConfig,
    offchainConfigVersion: flags.offchainConfigVersion,
    secret: flags.secret || env.secret,
    randomSecret: flags.randomSecret || undefined,
  }
}

const makeContractInput = async (
  input: SetConfigInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
  console.log('making contract input')
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

// TODO: beforeExecute attempt to deserialize offchainConfig to check for validity
// TODO: diff onchain config vs proposed config

const afterExecute: AfterExecute<SetConfigInput, ContractInput> = (context, input, deps) => async (
  result,
) => {
  const txHash = result.responses[0].tx.hash
  const txInfo = (await context.provider.provider.getTransactionReceipt(
    txHash,
  )) as InvokeTransactionReceiptResponse
  if (txInfo.status === 'REJECTED') {
    return { successfulConfiguration: false }
  }
  const eventData = txInfo.events[0].data

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
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
  hooks: {
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
