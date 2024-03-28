import Deploy from './deploy'
import SetL1Sender from './setL1Sender'
import Declare from './declare'
import Upgrade from './upgrade'
import Inspection from './inspection'

export const executionCommands = [Deploy, Declare, SetL1Sender, Upgrade]
export const inspectionCommands = [...Inspection]
