import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  OCR2 = 'aggregator',
  ACCESS_CONTROLLER = 'simple_write_access_controller',
  PROXY = 'aggregator_proxy',
  AGGREGATOR_CONSUMER = 'Aggregator_consumer',
}

export const loadContract_Ocr2 = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/starknet-artifacts/src/chainlink/cairo/ocr2/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadContract_Example = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/starknet-artifacts/src/chainlink/cairo/example/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadContract_AccessController = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/starknet-artifacts/src/chainlink/cairo/access/SimpleWriteAccessController/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const ocr2ContractLoader = () => loadContract_Ocr2(CONTRACT_LIST.OCR2)
export const ocr2ProxyLoader = () => loadContract_Ocr2(CONTRACT_LIST.PROXY)
export const aggregatorConsumerLoader = () =>
  loadContract_Example(CONTRACT_LIST.AGGREGATOR_CONSUMER)
export const accessControllerContractLoader = () =>
  loadContract_AccessController(CONTRACT_LIST.ACCESS_CONTROLLER)
