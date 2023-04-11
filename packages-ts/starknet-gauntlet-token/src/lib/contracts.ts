import fs from 'fs'
import { CompiledContract, json } from 'starknet'
import { ContractFactory } from 'ethers'

export enum CONTRACT_LIST {
  TOKEN = 'token',
}

// todo: remove when stable contract artifacts release available
const CONTRACT_NAME_TO_ARTIFACT_NAME = {
  [CONTRACT_LIST.TOKEN]: 'link_token',
}

export const loadTokenContract = (name: CONTRACT_LIST): CompiledContract => {
  const artifactName = CONTRACT_NAME_TO_ARTIFACT_NAME[name]
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/starknet-artifacts/src/chainlink/cairo/token/starkgate/presets/${artifactName}.cairo/${artifactName}.json`,
      )
      .toString('ascii'),
  )
}

export const loadOZContract = (name: CONTRACT_LIST): any => {
  const artifactName = CONTRACT_NAME_TO_ARTIFACT_NAME[name]
  const abi = json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@openzeppelin/contracts/build/contracts/${artifactName}.json`,
      )
      .toString('ascii'),
  )
  return new ContractFactory(abi?.abi, abi?.bytecode)
}

export const tokenContractLoader = () => loadTokenContract(CONTRACT_LIST.TOKEN)
