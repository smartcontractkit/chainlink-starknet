import fs from 'fs'
import { json } from 'starknet'

export enum CONTRACT_LIST {
  TOKEN = 'token',
}

// todo: remove when stable contract artifacts release available
const CONTRACT_NAME_TO_ARTIFACT_NAME = {
  [CONTRACT_LIST.TOKEN]: 'LinkToken',
}

export const loadTokenContract = (name: CONTRACT_LIST) => {
  const artifactName = CONTRACT_NAME_TO_ARTIFACT_NAME[name]
  return {
    contract: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${artifactName}.sierra.json`,
        'utf-8',
      ),
    ),
    casm: json.parse(
      fs.readFileSync(
        `${__dirname}/../../../../contracts/target/release/chainlink_${artifactName}.casm.json`,
        'utf-8',
      ),
    ),
  }
}

export const tokenContractLoader = () => loadTokenContract(CONTRACT_LIST.TOKEN)
