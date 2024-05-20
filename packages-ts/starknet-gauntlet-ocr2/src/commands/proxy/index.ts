import Deploy from './deploy'
import Declare from './declare'
import Upgrade from './upgrade'
import Inspect from './inspection/inspect'
import ProposeAggregator from './proposeAggregator'
import ConfirmAggregator from './confirmAggregator'
import TransferOwnership from './transferOwnership'
import AcceptOwnership from './acceptOwnership'

export const executeCommands = [
  Deploy,
  Declare,
  Upgrade,
  ProposeAggregator,
  ConfirmAggregator,
  TransferOwnership,
  AcceptOwnership,
]
export const inspectionCommands = [Inspect]
