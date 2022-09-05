import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  OCR2 = 'aggregator',
  ACCESS_CONTROLLER = 'simple_write_access_controller',
}

export const loadContractOcr2 = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starknet-contracts-ocr2/artifacts/ocr2/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadContractAccessController = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@chainlink-dev/starknet-contracts-ocr2/artifacts/access/SimpleWriteAccessController/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const ocr2ContractLoader = () => loadContractOcr2(CONTRACT_LIST.OCR2)
export const accessControllerContractLoader = () =>
  loadContractAccessController(CONTRACT_LIST.ACCESS_CONTROLLER)
