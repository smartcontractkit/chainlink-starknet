import Deploy from './deploy'
import inspect from './inspection/inspect'
import SetBilling from './setBilling'
import SetConfig from './setConfig'
import AddAccess from './addAccess'

export const executeCommands = [Deploy, AddAccess, SetBilling, SetConfig]
export const inspectionCommands = [inspect]
