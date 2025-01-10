import { loadContractByPath } from '@chainlink/starknet-gauntlet'

export enum CONTRACT_LIST {
  BRIDGE = 'bridge',
}

export const bridgeContractLoader = () =>
  loadContractByPath(
    `${__dirname}/../../artifacts/bridge/TokenBridge.contract_class.json`,
    `${__dirname}/../../artifacts/bridge/TokenBridge.compiled_contract_class.json`,
  )
