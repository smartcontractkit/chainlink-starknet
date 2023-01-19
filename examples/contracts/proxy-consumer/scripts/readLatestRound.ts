import fs from 'fs'
import {
  Account,
  Provider,
  Contract,
  CallContractResponse,
  json,
  ec,
  transaction,
  Status,
} from 'starknet'

// StarkNet network: Either goerli-alpha or mainnet-alpha
const network = 'goerli-alpha'

/** Environment variables for a deployed and funded account to use for deploying contracts
 * Find your OpenZeppelin account address and private key at:
 * ~/.starknet_accounts/starknet_open_zeppelin_accounts.json
 */
const accountAddress = process.env.DEPLOYER_ACCOUNT_ADDRESS as string
const accountKeyPair = ec.getKeyPair(process.env.DEPLOYER_PRIVATE_KEY as string)

const consumerContractName = 'Proxy_consumer'

const contractAddress = process.argv.at(2) as string

const provider = new Provider({
  sequencer: {
    network: network,
  },
})

const account = new Account(provider, accountAddress, accountKeyPair)

export async function updateStoredRound(account: Account, contractAddress: string) {
  const consumerContract = json.parse(
    fs
      .readFileSync(
        `${__dirname}/../starknet-artifacts/contracts/${consumerContractName}.cairo/${consumerContractName}.json`,
      )
      .toString('ascii'),
  )

  const targetContract = new Contract(consumerContract.abi, contractAddress, account)

  const response = await targetContract.invoke('get_latest_round_data')

  console.log('\nInvoking the get_latest_round_data function.')
  console.log('Transaction hash: ' + response.transaction_hash)

  console.log('Waiting for transaction...')
  let transactionStatus = (await provider.getTransactionReceipt(response.transaction_hash)).status
  while (transactionStatus !== 'REJECTED' && transactionStatus !== 'ACCEPTED_ON_L2') {
    console.log('Transaction status is: ' + transactionStatus)
    await new Promise((f) => setTimeout(f, 10000))
    transactionStatus = (await provider.getTransactionReceipt(response.transaction_hash)).status
  }
  console.log('Transaction is: ' + transactionStatus)
  readStoredRound(account, contractAddress)
}

export async function readStoredRound(account: Account, contractAddress: string) {
  const round = await account.callContract({
    contractAddress: contractAddress,
    entrypoint: 'get_stored_round',
  })

  console.log('\nStored values are:')
  printResult(round)
  return round
}

export async function readStoredProxy(account: Account, contractAddress: string) {
  const feed = await account.callContract({
    contractAddress: contractAddress,
    entrypoint: 'get_stored_feed_address',
  })

  return feed
}

function printResult(latestRound: CallContractResponse) {
  console.log('round_id =', parseInt(latestRound.result[0], 16))
  console.log('answer =', parseInt(latestRound.result[1], 16))
  console.log('block_num =', parseInt(latestRound.result[2], 16))
  console.log('observation_timestamp =', parseInt(latestRound.result[3], 16))
  console.log('transmission_timestamp =', parseInt(latestRound.result[4], 16))
}

updateStoredRound(account, contractAddress)
