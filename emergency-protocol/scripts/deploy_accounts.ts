import fs from 'fs'
import { starknet } from 'hardhat'

interface UserAccount {
  account: string
  privateKey: string
}
let firstAccount: UserAccount

async function main() {
  firstAccount = await createAccount()

  fs.appendFile(__dirname + '/.env', '\nACCOUNT_ADDRESS=' + firstAccount.account, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nPRIVATE_KEY=' + firstAccount.privateKey, function (err) {
    if (err) throw err
  })
}

async function createAccount(): Promise<UserAccount> {
  console.log('Deployment Tx - Account Contract to StarkNet...')
  const accountResponse = await starknet.deployAccount('OpenZeppelin')

  return { account: accountResponse.address as string, privateKey: accountResponse.privateKey }
}
main()
