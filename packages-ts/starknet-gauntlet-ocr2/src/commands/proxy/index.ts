import Deploy from './deploy'
import Declare from './declare'
import Inspect from './inspection/inspect'
import ProposeAggregator from './proposeAggregator'
import ConfirmAggregator from './confirmAggregator'

export const executeCommands = [Deploy, Declare, ProposeAggregator, ConfirmAggregator]
export const inspectionCommands = [Inspect]
