import Token from './token'
import L1Bridge from './L1-bridge'
import L2Bridge from './L2-bridge'
import Inspection from './inspection'

export const L1Commands = [...L1Bridge]
export const L2Commands = [...Token, ...L2Bridge]
export const InspectionCommands = [...Inspection]
