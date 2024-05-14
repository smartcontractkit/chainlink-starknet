import Deploy from './deploy'
import Declare from './declare'
import Upgrade from './upgrade'
import inspect from './inspection/inspect'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import SetPayees from './setPayees'
import AddAccess from './addAccess'
import DisableAccessCheck from './disableAccessCheck'
import ProposeOwnership from './proposeOwnership'
import AcceptOwnership from './acceptOwnership'

export const executeCommands = [
  Deploy,
  Declare,
  Upgrade,
  AddAccess,
  DisableAccessCheck,
  SetBilling,
  SetConfig,
  SetPayees,
  ProposeOwnership,
  AcceptOwnership,
]
export const inspectionCommands = [inspect]
