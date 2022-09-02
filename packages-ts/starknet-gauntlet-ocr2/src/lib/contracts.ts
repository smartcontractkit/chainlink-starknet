import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  OCR2 = 'aggregator',
  ACCESS_CONTROLLER = 'access_controller',
}

export const loadContract_Ocr2 = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starknet-contracts-ocr2/artifacts/ocr2/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadContract_AccessController = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starknet-contracts-ocr2/artifacts/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const ocr2ContractLoader = () => loadContract_Ocr2(CONTRACT_LIST.OCR2)
export const accessControllerContractLoader = () =>
  loadContract_AccessController(CONTRACT_LIST.ACCESS_CONTROLLER)
