import { existsSync, readFileSync } from 'fs'

export enum CONTRACT_TYPES {
  PROXY = 'proxies',
  ACCESS_CONTROLLER = 'accessControllers',
  AGGREGATOR = 'contracts',
  VALIDATOR = 'validators',
}

export const getRDD = (path: string): any => {
  // test whether the file exists as a relative path or an absolute path
  if (!existsSync(path)) {
    throw new Error(`Could not find the RDD. Make sure you provided a valid $path`)
  }

  try {
    return JSON.parse(readFileSync(path, 'utf8'))
  } catch (e) {
    throw new Error(`An error ocurred while parsing the RDD. Make sure you provided a valid path`)
  }
}
