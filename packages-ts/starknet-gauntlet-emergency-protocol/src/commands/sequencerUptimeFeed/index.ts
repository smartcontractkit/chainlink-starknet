import Deploy from './deploy'
import SetL1Sender from './setL1Sender'
import Upgrade from './upgrade'

import Inspection from './inspection'

export const executionCommands = [Deploy, SetL1Sender, Upgrade]
export const inspectionCommands = [...Inspection]
