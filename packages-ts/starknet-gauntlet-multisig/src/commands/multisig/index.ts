import Deploy from './deploy'
import SetSigners from './setSigners'
import SetThreshold from './setThreshold'

import Inspection from './inspection'

export const executeCommands = [Deploy, SetSigners, SetThreshold]
export const inspectionCommands = [...Inspection]
