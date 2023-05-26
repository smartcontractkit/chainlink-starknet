import { existsSync, readFileSync } from 'fs'
import { time } from '@chainlink/gauntlet-core/dist/utils'
import { SetConfigInput } from '@chainlink/gauntlet-contracts-ocr2'

export const getRDD = (path: string): any => {
  path = path || process.env.RDD
  if (!path) {
    throw new Error(
      `No reference data directory specified!  Must pass in the '--rdd' flag or set the 'RDD' env var`,
    )
  }

  let pathToUse
  if (existsSync(path)) {
    pathToUse = path
  } else {
    throw new Error(`Could not find the RDD file. Make sure you provided a valid RDD path.`)
  }

  try {
    const buffer = readFileSync(pathToUse, 'utf8')
    return JSON.parse(buffer.toString())
  } catch (e) {
    throw new Error(
      `An error ocurred while parsing the RDD file. Make sure you provided a valid RDD path.`,
    )
  }
}

export const getConfigForContract = (rdd: any, contract: string): SetConfigInput => {
  const rddContract = rdd['contracts']?.[contract]
  if (!rddContract) {
    throw new Error(
      `Could not find the contract ${contract} in the RDD file. Make sure you provided a valid contract address.`,
    )
  }

  const config = rddContract['config']
  const operators: any[] = rddContract.oracles.map((o) => rdd.operators[o.operator])
  const operatorsPublicKeys = operators.map((o) => o.ocr2OffchainPublicKey[0])
  const operatorsPeerIds = operators.map((o) => o.peerId[0])
  const operatorConfigPublicKeys = operators.map((o) => o.ocr2ConfigPublicKey[0])
  const signers = operators.map((o) => o.ocr2OnchainPublicKey[0])
  const transmitters = operators.map((o) => o.ocrNodeAddress[0])

  return {
    f: config.f,
    signers,
    transmitters,
    onchainConfig: [],
    offchainConfig: {
      deltaProgressNanoseconds: time.durationToNanoseconds(config.deltaProgress).toNumber(),
      deltaResendNanoseconds: time.durationToNanoseconds(config.deltaResend).toNumber(),
      deltaRoundNanoseconds: time.durationToNanoseconds(config.deltaRound).toNumber(),
      deltaGraceNanoseconds: time.durationToNanoseconds(config.deltaGrace).toNumber(),
      deltaStageNanoseconds: time.durationToNanoseconds(config.deltaStage).toNumber(),
      rMax: config.rMax,
      s: config.s,
      offchainPublicKeys: operatorsPublicKeys,
      peerIds: operatorsPeerIds,
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
    },
    offchainConfigVersion: 2,
    secret: 'super mega secret', // todo: secret from env or randomized
  }
}
