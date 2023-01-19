import fs from 'fs'
import dotenv from 'dotenv'
import { CompiledContract, json, ec, Account, Provider } from 'starknet'

const DEVNET_NAME = 'devnet'

export const loadContract = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(`${__dirname}/../starknet-artifacts/contracts/${name}.cairo/${name}.json`)
      .toString('ascii'),
  )
}

export const loadContractPath = (path: string, name: string): CompiledContract => {
  return json.parse(
    fs.readFileSync(`${__dirname}/${path}/${name}.cairo/${name}.json`).toString('ascii'),
  )
}

export const loadContract_Account = (name: string): CompiledContract => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../node_modules/@shardlabs/starknet-hardhat-plugin/dist/contract-artifacts/OpenZeppelinAccount/0.5.1/${name}.cairo/${name}.json`,
      )
      .toString('ascii'),
  )
}

export const loadContract_Solidity = (path: string, name: string): any => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/artifacts/src/chainlink/solidity/${path}/${name}.sol/${name}.json`,
      )
      .toString('ascii'),
  )
}
export const loadContract_Solidity_V8 = (name: string): any => {
  return json.parse(
    fs
      .readFileSync(
        `${__dirname}/../../../../contracts/artifacts/@chainlink/contracts/src/v0.8/interfaces/${name}.sol/${name}.json`,
      )
      .toString('ascii'),
  )
}

export function createDeployerAccount(provider: Provider): Account {
  dotenv.config({ path: __dirname + '/../.env' })

  const privateKey: string = process.env.DEPLOYER_PRIVATE_KEY as string
  const accountAddress: string = process.env.DEPLOYER_ACCOUNT_ADDRESS as string
  if (!privateKey || !accountAddress) {
    throw new Error('Deployer account address or private key is undefined!')
  }

  const deployerKeyPair = ec.getKeyPair(privateKey)
  return new Account(provider, accountAddress, deployerKeyPair)
}

export const makeProvider = () => {
  const network = process.env.NETWORK || DEVNET_NAME
  if (network === DEVNET_NAME) {
    return new Provider({
      sequencer: {
        baseUrl: 'http://127.0.0.1:5050/',
        feederGatewayUrl: 'feeder_gateway',
        gatewayUrl: 'gateway',
      },
    })
  } else {
    return new Provider({
      sequencer: {
        network: 'goerli-alpha',
      },
    })
  }
}
