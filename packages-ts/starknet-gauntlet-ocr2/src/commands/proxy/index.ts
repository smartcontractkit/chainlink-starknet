import Deploy from './deploy'
import Inspect from './inspection/inspect'
import ProposeAggregator from './proposeAggregator'
import ConfirmAggregator from './confirmAggregator'

export const executeCommands = [Deploy, ProposeAggregator, ConfirmAggregator]
export const inspectionCommands = [Inspect]
