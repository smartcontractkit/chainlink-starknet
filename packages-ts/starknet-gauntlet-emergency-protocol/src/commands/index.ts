import UptimeFeedCommands from './sequencerUptimeFeed'
import StarknetValidatorCommands from './starknetValidator'

export const L1Commands = [...StarknetValidatorCommands]
export const L2Commands = [...UptimeFeedCommands]
