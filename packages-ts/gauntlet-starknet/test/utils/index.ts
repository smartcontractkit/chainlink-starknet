import { CompiledContract, json } from 'starknet'
import fs from 'fs'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../__mocks__/${name}.json`).toString('ascii'))
}

export const loadExampleContract = () => loadContract('example')

const noop = () => {}

export const noopLogger: typeof logger = {
  table: noop,
  log: noop,
  info: noop,
  warn: noop,
  success: noop,
  error: noop,
  loading: noop,
  line: noop,
  style: () => '',
  debug: noop,
  time: noop,
}

export const noopPrompt: typeof prompt = async () => {}
