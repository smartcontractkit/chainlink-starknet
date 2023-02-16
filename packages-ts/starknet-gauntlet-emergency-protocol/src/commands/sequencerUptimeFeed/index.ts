import Deploy from './deploy'
import SetL1Sender from './setL1Sender'

import Inspection from './inspection'

export const executionCommands = [Deploy, SetL1Sender]
export const inspectionCommands = [...Inspection]
