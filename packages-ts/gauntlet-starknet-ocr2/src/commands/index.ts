// import AccessController from './accessController'

import { executeCommands as acExecuteCommands } from './accessController'
import { executeCommands as ocr2ExecuteCommands } from './ocr2'

export const executeCommands = [...acExecuteCommands, ...ocr2ExecuteCommands]
