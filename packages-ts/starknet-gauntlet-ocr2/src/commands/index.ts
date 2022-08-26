// import AccessController from './accessController'

import { executeCommands as acExecuteCommands } from './accessController'
import {
  executeCommands as ocr2ExecuteCommands,
  inspectionCommands as ocr2InspectionCommands,
} from './ocr2'

export const executeCommands = [...acExecuteCommands, ...ocr2ExecuteCommands]
export const inspectionCommands = [...ocr2InspectionCommands]
