import Deploy from './deploy'
import Declare from './declare'
import Upgrade from './upgrade'
import Inspect from './inspection/inspect'
import ProposeAggregator from './proposeAggregator'
import ConfirmAggregator from './confirmAggregator'
import TransferOwnership from './transferOwnership'
import AcceptOwnership from './acceptOwnership'
import DisableAccessCheck from './disableAccessCheck'

export const executeCommands = [
  Deploy,
  Declare,
  Upgrade,
  ProposeAggregator,
  ConfirmAggregator,
  TransferOwnership,
  AcceptOwnership,
  DisableAccessCheck,
]
export const inspectionCommands = [Inspect]
