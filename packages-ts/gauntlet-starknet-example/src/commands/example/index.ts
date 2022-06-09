import Deploy from './deploy'
import IncreaseBalance from './increaseBalance'
import Inspection from './inspection'
export const executeCommands = [Deploy, IncreaseBalance]
export const inspectionCommands = [...Inspection]
