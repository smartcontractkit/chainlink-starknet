import Deploy from './deploy'
import SetL1Sender from './setL1Sender'
import Declare from './declare'

import Inspection from './inspection'

export const executionCommands = [Deploy, SetL1Sender, Declare]
export const inspectionCommands = [...Inspection]
