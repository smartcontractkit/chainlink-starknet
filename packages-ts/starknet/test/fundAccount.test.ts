import { assert } from 'chai'
import { account } from '@chainlink/starknet'
import { Account, ec, SequencerProvider, stark } from 'starknet'
import { DEVNET_URL, ERC20_ADDRESS_DEVNET, ERC20_ADDRESS_TESTNET } from '../src/utils'

describe('fundAccount', function () {
  this.timeout(900_000)
  let alice: Account
  let bob: Account
  let provider: SequencerProvider
  let erc20_address: string

  const opts = account.makeFunderOptsFromEnv()
  const funder = account.Funder(opts)

  before(async function () {
    const gateway_url = process.env.NODE_URL || DEVNET_URL
    provider = new SequencerProvider({ baseUrl: gateway_url })

    if (gateway_url === DEVNET_URL) {
      erc20_address = ERC20_ADDRESS_DEVNET
    } else {
      erc20_address = ERC20_ADDRESS_TESTNET
    }

    const alice_starkKeyPair = ec.genKeyPair()
    const bob_starkKeyPair = ec.genKeyPair()

    const default_alice_address = stark.randomAddress()
    const default_bob_address = stark.randomAddress()

    alice = new Account(provider, default_alice_address, alice_starkKeyPair)
    bob = new Account(provider, default_bob_address, bob_starkKeyPair)

    await funder.fund([
      { account: alice.address, amount: 5000 },
      { account: bob.address, amount: 8000 },
    ])
  })

  it('should have fund alice', async () => {
    const balance = await alice.callContract({
      contractAddress: erc20_address,
      entrypoint: 'balanceOf',
      calldata: [BigInt(alice.address).toString(10)],
    })
    assert.deepEqual(balance.result, ['0x1388', '0x0'])
  })

  it('should have fund bob', async () => {
    const balance = await bob.callContract({
      contractAddress: erc20_address,
      entrypoint: 'balanceOf',
      calldata: [BigInt(bob.address).toString(10)],
    })
    assert.deepEqual(balance.result, ['0x1f40', '0x0'])
  })

  it("should increament alice's fees", async () => {
    await funder.fund([{ account: alice.address, amount: 100 }])

    const balance = await alice.callContract({
      contractAddress: erc20_address,
      entrypoint: 'balanceOf',
      calldata: [BigInt(alice.address).toString(10)],
    })
    assert.deepEqual(balance.result, ['0x13ec', '0x0'])
  })

  it("should increament bob's fees", async () => {
    await funder.fund([{ account: bob.address, amount: 1000 }])

    const balance = await bob.callContract({
      contractAddress: erc20_address,
      entrypoint: 'balanceOf',
      calldata: [BigInt(bob.address).toString(10)],
    })
    assert.deepEqual(balance.result, ['0x2328', '0x0'])
  })
})
