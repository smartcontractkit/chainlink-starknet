import fs from 'fs'
import { CompiledContract, json } from 'starknet'

export enum CONTRACT_LIST {
  OCR2 = 'Aggregator',
  ACCESS_CONTROLLER = 'SimpleWriteAccessController',
  PROXY = 'AggregatorProxy',
  AGGREGATOR_CONSUMER = 'AggregatorConsumer',
}

export const loadContract = (name: CONTRACT_LIST) => {
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.sierra.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${name}.casm.json`,
        'utf-8',
      ),
    ),
  }
}

export const ocr2ContractLoader = () => loadContract(CONTRACT_LIST.OCR2)
export const ocr2ProxyLoader = () => loadContract(CONTRACT_LIST.PROXY)
export const aggregatorConsumerLoader = () => loadContract(CONTRACT_LIST.AGGREGATOR_CONSUMER)
export const accessControllerContractLoader = () => loadContract(CONTRACT_LIST.ACCESS_CONTROLLER)
