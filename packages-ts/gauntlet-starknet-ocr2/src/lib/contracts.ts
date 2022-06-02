import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  OCR2 = 'aggregator',
  ACCESS_CONTROLLER = 'access_controller',
}

export const loadContract = (name: CONTRACT_LIST): CompiledContract => {
  return json.parse(fs.readFileSync(`${__dirname}/../../artifacts/abi/${name}.json`).toString('ascii'))
}

export const ocr2ContractLoader = () => loadContract(CONTRACT_LIST.OCR2)
export const accessControllerContractLoader = () => loadContract(CONTRACT_LIST.ACCESS_CONTROLLER)
