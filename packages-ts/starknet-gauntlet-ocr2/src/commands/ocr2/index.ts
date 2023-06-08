import Deploy from './deploy'
import Upgrade from './upgrade'
import inspect from './inspection/inspect'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import AddAccess from './addAccess'
import DisableAccessCheck from './disableAccessCheck'

export const executeCommands = [
  Deploy,
  Upgrade,
  AddAccess,
  DisableAccessCheck,
  SetBilling,
  SetConfig,
]
export const inspectionCommands = [inspect]
