// import AccessController from './accessController'

import { executeCommands as acExecuteCommands } from './accessController'
import {
  executeCommands as ocr2ExecuteCommands,
  inspectionCommands as ocr2InspectionCommands,
} from './ocr2'
import { executeCommands as aggregatorProxyExecuteCommands } from './aggregatorProxy'

export const executeCommands = [
  ...acExecuteCommands,
  ...ocr2ExecuteCommands,
  ...aggregatorProxyExecuteCommands,
]
export const inspectionCommands = [...ocr2InspectionCommands]
