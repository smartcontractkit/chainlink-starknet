import Deploy from './deploy'
import SetOwners from './setOwners'
import SetThreshold from './setThreshold'

import Inspection from './inspection'

export const executeCommands = [Deploy, SetOwners, SetThreshold]
export const inspectionCommands = [...Inspection]
