import Deploy from './deploy'
import Upgrade from './upgrade'
import Inspect from './inspection/inspect'
import ProposeAggregator from './proposeAggregator'
import ConfirmAggregator from './confirmAggregator'

export const executeCommands = [Deploy, Upgrade, ProposeAggregator, ConfirmAggregator]
export const inspectionCommands = [Inspect]
