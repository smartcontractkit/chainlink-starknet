import { starknet, type } from 'hardhat'
import { assert } from 'chai'
import { account } from '@chainlink/starknet'
import fs from 'fs'
import { readDecimals } from '../../scripts/readDecimals'
import { readLatestRound } from '../../scripts/readLatestRound'
import { createDeployerAccount, loadContract, makeProvider } from '../../scripts/utils'
import { Contract, Provider, number } from 'starknet'
import dotenv from 'dotenv'
import { deployContract } from '../../scripts/deploy_contracts'
import { getLatestPrice } from '../../scripts/getLatestPrice'
import { deployAccount } from '../../scripts/deploy_accounts'
import { consumerValidator } from '../../scripts/consumerValidator'

describe('ExamplesTests', function () {
  dotenv.config({ path: __dirname + '/../../.env' })

  this.timeout(600_000)
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let alice: type.Account
  let provider: Provider

  before(async () => {
    provider = makeProvider()

    alice = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([{ account: alice.address, amount: 1e21 }])
    await alice.deployAccount()
    fs.appendFile(
      __dirname + '/../../.env',
      '\nDEPLOYER_ACCOUNT_ADDRESS=' + alice.address,
      function (err) {
        if (err) throw err
      },
    )
    fs.appendFile(
      __dirname + '/../../.env',
      '\nDEPLOYER_PRIVATE_KEY=' + alice.privateKey,
      function (err) {
        if (err) throw err
      },
    )
  })

  it('should deploy contract', async () => {
    await deployContract()
  })

  it('should deploy account', async () => {
    await deployAccount()
  })

  it('should set and read latest round data successfully', async () => {
    const MockArtifact = loadContract('MockAggregator')
    const mock = new Contract(MockArtifact.abi, process.env.MOCK as string, provider)

    const bob = createDeployerAccount(provider)
    await funder.fund([{ account: bob.address, amount: 1e21 }])
    const transaction = await bob.execute(
      {
        contractAddress: mock.address,
        entrypoint: 'set_latest_round_data',
        calldata: [42, 3, 9876, 27839],
      },
      [mock.abi],
    )
    console.log('Waiting for Tx to be Accepted on Starknet - Aggregator consumer Deployment...')
    await provider.waitForTransaction(transaction.transaction_hash)

    const latestRound = await readLatestRound()
    assert.equal(parseInt(latestRound.result[1], 16), 42)
    assert.equal(parseInt(latestRound.result[2], 16), 3)
    assert.equal(parseInt(latestRound.result[3], 16), 9876)
    assert.equal(parseInt(latestRound.result[4], 16), 27839)
  })

  it('should read Decimals successfully', async () => {
    const decimals = await readDecimals()
    assert.equal(decimals, 18)
  })

  it('should test consumer validator', async () => {
    await consumerValidator()
  })

  it('should get latest price', async () => {
    const latestPrice = await getLatestPrice()
    assert.equal(parseInt(latestPrice.result[0], 16), 42)
  })
})
