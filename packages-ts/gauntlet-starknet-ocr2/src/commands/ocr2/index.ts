import Deploy from './deploy'
import inspect from './inspection/inspect'
import SetBilling from './setBilling'
import SetConfig from './setConfig'

export const executeCommands = [Deploy, SetBilling, SetConfig]
export const inspectionCommands = [inspect]
