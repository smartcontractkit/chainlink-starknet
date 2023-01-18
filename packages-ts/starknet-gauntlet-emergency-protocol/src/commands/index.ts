import {
    executionCommands as UptimeFeedCommands,
    inspectionCommands as UptimeInspectionCommands
} from './sequencerUptimeFeed'
import StarknetValidatorCommands from './starknetValidator'

export const L1Commands = [...StarknetValidatorCommands]
export const L2Commands = [...UptimeFeedCommands]
export const L2InspectionCommands = [...UptimeFeedCommands]
