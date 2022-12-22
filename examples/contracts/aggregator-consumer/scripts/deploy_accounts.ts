import fs from 'fs'
import dotenv from 'dotenv'

import { defaultProvider, ec, stark, Account, Provider } from 'starknet'
import { loadContract_Account } from './index'

dotenv.config({ path: __dirname + '/../.env' })

const ACCOUNT_NAME = 'Account'
interface UserAccount {
  account: string
  privateKey: string
}
let firstAccount: UserAccount
let secondAccount: UserAccount

async function main() {
  firstAccount = await createAccount()
  secondAccount = await createAccount()

  fs.appendFile(__dirname + '/.env', '\nACCOUNT_ADDRESS=' + firstAccount.account, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nPRIVATE_KEY=' + firstAccount.privateKey, function (err) {
    if (err) throw err
  })

  fs.appendFile(__dirname + '/.env', '\nACCOUNT_ADDRESS_2=' + secondAccount.account, function (
    err,
  ) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nPRIVATE_KEY_2=' + secondAccount.privateKey, function (err) {
    if (err) throw err
  })
}

function createDeployerAccount(provider: Provider): Account {
  const privateKey: string = process.env.DEPLOYER_PRIVATE_KEY as string
  const accountAddress: string = process.env.DEPLOYER_ACCOUNT_ADDRESS as string
  if (!privateKey || !accountAddress) {
    throw new Error('Deployer account address or private key is undefined!')
  }

  const deployerKeyPair = ec.getKeyPair(privateKey)
  return new Account(provider, accountAddress, deployerKeyPair)
}

async function createAccount(): Promise<UserAccount> {
  const compiledAccount = loadContract_Account(ACCOUNT_NAME)
  const privateKey = stark.randomAddress()

  const starkKeyPair = ec.getKeyPair(privateKey)
  const starkKeyPub = ec.getStarkKey(starkKeyPair)

  console.log('Deployment Tx - Account Contract to StarkNet...')
  const deployerAccount = createDeployerAccount(defaultProvider)
  const declareDeployResponse = await deployerAccount.declareDeploy({
    classHash: '0x0750cd490a7cd1572411169eaa8be292325990d33c5d4733655fe6b926985062', // OZ Account 0.5.0 class hash
    contract: compiledAccount,
    constructorCalldata: [starkKeyPub],
  })

  console.log('Waiting for Tx to be Accepted on Starknet - OZ Account Deployment...')
  const accountResponse = declareDeployResponse.deploy
  await defaultProvider.waitForTransaction(accountResponse.transaction_hash)

  return { account: accountResponse.address as string, privateKey: privateKey }
}

main()
