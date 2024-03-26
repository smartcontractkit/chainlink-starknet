import Deploy from './deploy'
import Declare from './declare'
import SetSigners from './setSigners'
import SetThreshold from './setThreshold'
import Upgrade from './upgrade'

import Inspection from './inspection'

export const executeCommands = [Deploy, Declare, SetSigners, SetThreshold, Upgrade]
export const inspectionCommands = [...Inspection]
