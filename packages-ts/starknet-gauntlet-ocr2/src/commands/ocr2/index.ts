import Deploy from './deploy'
import inspect from './inspection/inspect'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import AddAccess from './addAccess'
import DisableAccessCheck from './disableAccessCheck'

export const executeCommands = [Deploy, AddAccess, DisableAccessCheck, SetBilling, SetConfig]
export const inspectionCommands = [inspect]
