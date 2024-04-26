import fs from 'fs'
import {
  CONTRACT_TYPES,
  AfterExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  bytesToFelts,
  BeforeExecute,
  getRDD,
} from '@chainlink/starknet-gauntlet'
import { time, diff } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { SetConfig, encoding, SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'
import { decodeOffchainConfigFromEventData } from '../../lib/encoding'
import assert from 'assert'
import { getLatestOCRConfigEvent } from './inspection/configEvent'
import { BigNumberish, GetTransactionReceiptResponse } from 'starknet'

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
    const peerIds = operators.map((o) => o.peerId[0])
    const offchainConfig: encoding.OffchainConfig = {
      deltaProgressNanoseconds: time.durationToNanoseconds(config.deltaProgress).toNumber(),
      deltaResendNanoseconds: time.durationToNanoseconds(config.deltaResend).toNumber(),
      deltaRoundNanoseconds: time.durationToNanoseconds(config.deltaRound).toNumber(),
      deltaGraceNanoseconds: time.durationToNanoseconds(config.deltaGrace).toNumber(),
      deltaStageNanoseconds: time.durationToNanoseconds(config.deltaStage).toNumber(),
      rMax: config.rMax,
      s: config.s,
      offchainPublicKeys: operators.map((o) => o.ocr2OffchainPublicKey[0]),
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
      configPublicKeys: operators.map((o) => o.ocr2ConfigPublicKey[0]),
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
    f: parseInt(flags.f),
    signers: flags.signers,
    transmitters: flags.transmitters,
    onchainConfig: flags.onchainConfig,
    offchainConfig: flags.offchainConfig,
    offchainConfigVersion: parseInt(flags.offchainConfigVersion),
    secret: flags.secret || env.secret,
    randomSecret: flags.randomSecret || undefined,
  }
}

const makeContractInput = async (
  input: SetConfigInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
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
  const onchainConfig = [] // onchain config should be empty array for input (generate onchain)
  return [oracles, input.f, onchainConfig, 2, bytesToFelts(offchainConfig)]
}

const beforeExecute: BeforeExecute<SetConfigInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.loading(`Executing ${context.id} from contract ${context.contractAddress}`)

  const { offchainConfig } = await encoding.serializeOffchainConfig(
    input.user.offchainConfig,
    input.user.secret,
  )

  const newOffchainConfig = encoding.deserializeConfig(offchainConfig)

  const rawEvents = await getLatestOCRConfigEvent(context.provider, context.contractAddress)
  if (rawEvents.length === 0) {
    // if no config set events found in the given block, throw error
    // this should not happen if block number in latestConfigDetails is set correctly
    deps.logger.info('No previous config found, review the offchain config below:')
    deps.logger.log(newOffchainConfig)
    return
  }
  // assume last event found is the latest config, in the event that multiple
  // set_config transactions ended up in the same block
  const events = context.contract.parseEvents({
    events: rawEvents,
  } as GetTransactionReceiptResponse)
  const event = events[events.length - 1]['ConfigSet']
  const currOffchainConfig = decodeOffchainConfigFromEventData(
    event.offchain_config as BigNumberish[],
  )

  deps.logger.info(
    'Review the proposed offchain config changes below: green - added, red - deleted.',
  )
  diff.printDiff(currOffchainConfig, newOffchainConfig)
}

const afterExecute: AfterExecute<SetConfigInput, ContractInput> = (context, input, deps) => async (
  result,
) => {
  const txHash = result.responses[0].tx.hash
  const txInfo = await context.provider.provider.getTransactionReceipt(txHash)
  if (!txInfo.isSuccess()) {
    return { successfulConfiguration: false }
  }
  const events = context.contract.parseEvents(txInfo)
  const event = events[events.length - 1]['ConfigSet']
  const offchainConfig = decodeOffchainConfigFromEventData(event.offchain_config as BigNumberish[])

  try {
    // remove cfg keys from user input
    delete input.user.offchainConfig.configPublicKeys
    assert.deepStrictEqual(offchainConfig, input.user.offchainConfig)
    deps.logger.success('Configuration was successfully set')

    // write lastConfigDigest back to RDD
    const configDigest = `0x${(event.latest_config_digest as bigint).toString(16)}`
    deps.logger.info(`lastConfigDigest to save in RDD: ${configDigest}`)
    if (context.flags.rdd) {
      deps.logger.info(`rdd file found! will automatically lastConfigDigest for you`)
      const rdd = getRDD(context.flags.rdd)
      const newRdd = { ...rdd, ...{ 'lastConfigDigest': configDigest } }
      fs.writeFileSync(context.flags.rdd, JSON.stringify(newRdd, null, 2))
      deps.logger.info(`rdd file ${context.flags.rdd} updated. please reformat file`)
    } else {
      deps.logger.info(`You must manually update lastConfigDigest yourself`)
    }

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
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
