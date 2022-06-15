import fs from 'fs'

import { defaultProvider, ec, stark } from 'starknet'
import { loadAccount } from './index'

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

  fs.appendFile(__dirname + '/.env', '\nACCOUNT_ADDRESS_2=' + secondAccount.account, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nPRIVATE_KEY_2=' + secondAccount.privateKey, function (err) {
    if (err) throw err
  })
}

async function createAccount(): Promise<UserAccount> {
  const compiledAccount = loadAccount(ACCOUNT_NAME)
  const privateKey = stark.randomAddress()

  const starkKeyPair = ec.getKeyPair(privateKey)
  const starkKeyPub = ec.getStarkKey(starkKeyPair)

  console.log('Deployment Tx - Account Contract to StarkNet...')
  const accountResponse = await defaultProvider.deployContract({
    contract: compiledAccount,
    constructorCalldata: [starkKeyPub],
  })

  console.log('Waiting for Tx to be Accepted on Starknet - Argent Account Deployment...')
  await defaultProvider.waitForTransaction(accountResponse.transaction_hash)

  return { account: accountResponse.address as string, privateKey: privateKey }
}
main()
