import { assert } from 'chai'
import { starknet } from 'hardhat'
import { Account } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'
import { AccountFunder } from '@chainlink/starknet/src/utils'

describe('fundAccount', function () {
  this.timeout(TIMEOUT)
  let alice: Account
  let bob: Account
  let funder: AccountFunder

  before(async function () {
    const opts = { network: 'devnet' }
    funder = new AccountFunder(opts)
    alice = await starknet.deployAccount('OpenZeppelin')
    bob = await starknet.deployAccount('OpenZeppelin')
    await funder.fund([
      { account: alice.address, amount: 5000 },
      { account: bob.address, amount: 8000 },
    ])
  })

  it('should have fund alice', async () => {
    let gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'
    const response = await fetch(`${gateway_url}/account_balance?address=${alice.address}`, {
      method: 'get',
      headers: { 'Content-Type': 'application/json' },
    })

    const data = await response.json()
    assert.equal(data.amount, 5000)
    assert.equal(data.unit, 'wei')
  })

  it('should have fund bob', async () => {
    let gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'
    const response = await fetch(`${gateway_url}/account_balance?address=${bob.address}`, {
      method: 'get',
      headers: { 'Content-Type': 'application/json' },
    })

    const data = await response.json()
    assert.equal(data.amount, 8000)
    assert.equal(data.unit, 'wei')
  })

  it("should increament alice's fees", async () => {
    await funder.fund([{ account: alice.address, amount: 100 }])

    let gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'
    const response = await fetch(`${gateway_url}/account_balance?address=${alice.address}`, {
      method: 'get',
      headers: { 'Content-Type': 'application/json' },
    })

    const data = await response.json()
    assert.equal(data.amount, 5100)
    assert.equal(data.unit, 'wei')
  })

  it("should increament bob's fees", async () => {
    await funder.fund([{ account: bob.address, amount: 1000 }])

    let gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'
    const response = await fetch(`${gateway_url}/account_balance?address=${bob.address}`, {
      method: 'get',
      headers: { 'Content-Type': 'application/json' },
    })

    const data = await response.json()
    assert.equal(data.amount, 9000)
    assert.equal(data.unit, 'wei')
  })
})
