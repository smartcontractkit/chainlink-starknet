import {
  executionCommands as UptimeFeedCommands,
  inspectionCommands as UptimeInspectionCommands,
} from './sequencerUptimeFeed'
import StarknetValidatorCommands from './L1StarknetValidator'
import L1AccessController from './L1AccessController'
import L1GasPriceFeed from './L1GasPriceFeed'

export const L1Commands = [...StarknetValidatorCommands, ...L1AccessController, ...L1GasPriceFeed]
export const L2Commands = [...UptimeFeedCommands]
export const L2InspectionCommands = [...UptimeInspectionCommands]
