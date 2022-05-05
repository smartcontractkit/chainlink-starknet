import OCR2Commands from '@chainlink/gauntlet-starknet-ocr2'
import ExampleCommands from '@chainlink/gauntlet-starknet-example'
import { executeCLI } from '@chainlink/gauntlet-core'
import { existsSync } from 'fs'
import path from 'path'
import { io } from '@chainlink/gauntlet-core/dist/utils'

const commands = {
  custom: [...OCR2Commands, ...ExampleCommands],
  loadDefaultFlags: () => ({}),
  abstract: {
    findPolymorphic: () => undefined,
    makeCommand: () => undefined,
  },
}

;(async () => {
  try {
    const networkPossiblePaths = ['./packages/gauntlet-cli/networks']
    const networkPath = networkPossiblePaths.filter((networkPath) =>
      existsSync(path.join(process.cwd(), networkPath)),
    )[0]
    const result = await executeCLI(commands, networkPath)
    if (result) {
      io.saveJSON(result, process.env['REPORT_NAME'] ? process.env['REPORT_NAME'] : 'report')
    }
    process.exit(0)
  } catch (e) {
    console.log('Starknet Command execution error', e.message)
    process.exitCode = 1
  }
})()
