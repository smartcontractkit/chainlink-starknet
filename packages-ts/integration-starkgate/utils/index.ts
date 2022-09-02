import fs from 'fs'
import { json } from 'starknet'

export const loadStarkgateContract = (name: string): any => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../node_modules/internals-starkgate-contracts/artifacts/0.0.3/eth/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadOpenzepplinContract = (name: string): any => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../node_modules/@openzeppelin/contracts/build/contracts/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadSolidityContract = (name: string): any => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../contracts/artifacts/src/chainlink/solidity/mocks/${name}.sol/${name}.json`,
      )
      .toString('ascii'),
  )
}
