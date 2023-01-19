// import AccessController from './accessController'

import { executeCommands as acExecuteCommands } from './accessController'
import {
  executeCommands as ocr2ExecuteCommands,
  inspectionCommands as ocr2InspectionCommands,
} from './ocr2'
import {
  executeCommands as proxyExecuteCommands,
  inspectionCommands as proxyInspectionCommands,
} from './proxy'
import {
  executeCommands as exampleExecuteCommands,
  inspectionCommands as exampleInspectionCommands,
} from './example'

export const executeCommands = [
  ...acExecuteCommands,
  ...ocr2ExecuteCommands,
  ...proxyExecuteCommands,
  ...exampleExecuteCommands,
]
export const inspectionCommands = [
  ...ocr2InspectionCommands,
  ...proxyInspectionCommands,
  ...exampleInspectionCommands,
]
