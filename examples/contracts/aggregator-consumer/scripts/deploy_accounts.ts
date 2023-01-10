import fs from 'fs'
import dotenv from 'dotenv'
import { ec, stark } from 'starknet'
import { loadContract_Account, createDeployerAccount, makeProvider } from './utils'

dotenv.config({ path: __dirname + '/../.env' })

const ACCOUNT_NAME = 'Account'
interface UserAccount {
  account: string
  privateKey: string
}
let firstAccount: UserAccount

export async function deployAccount() {
  firstAccount = await createAccount()

  fs.appendFile(__dirname + '/../.env', '\nACCOUNT_ADDRESS=' + firstAccount.account, function (
    err,
  ) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/../.env', '\nPRIVATE_KEY=' + firstAccount.privateKey, function (err) {
    if (err) throw err
  })
}

async function createAccount(): Promise<UserAccount> {
  const provider = makeProvider()

  const predeployedAccount = createDeployerAccount(provider)

  const compiledAccount = loadContract_Account(ACCOUNT_NAME)
  const privateKey = stark.randomAddress()

  const starkKeyPair = ec.getKeyPair(privateKey)
  const starkKeyPub = ec.getStarkKey(starkKeyPair)
  const OZaccountConstructorCallData = stark.compileCalldata({ publicKey: starkKeyPub })

  const declareTx = await predeployedAccount.declare({
    classHash: '0x4d07e40e93398ed3c76981e72dd1fd22557a78ce36c0515f679e27f0bb5bc5f',
    contract: compiledAccount,
  })

  console.log('Declare new Account...')
  await provider.waitForTransaction(declareTx.transaction_hash)

  const salt = '900080545022'
  const accountResponse = await predeployedAccount.deploy({
    classHash: '0x4d07e40e93398ed3c76981e72dd1fd22557a78ce36c0515f679e27f0bb5bc5f',
    constructorCalldata: OZaccountConstructorCallData,
    salt,
  })

  console.log('Waiting for Tx to be Accepted on Starknet - OZ Account Deployment...')
  await provider.waitForTransaction(accountResponse.transaction_hash)

  return { account: accountResponse.contract_address[0], privateKey: privateKey }
}

deployAccount()
