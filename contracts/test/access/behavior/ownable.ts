import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account } from 'hardhat/types/runtime'

export type BeforeFn = () => Promise<TestData>
export type TestData = {
  ownable: StarknetContract
  alice: Account
  bob: Account
}

// NOTICE: Leading zeros are trimmed for an encoded felt (number).
//   To decode, the raw felt needs to be start padded up to max felt size (252 bits or < 32 bytes).
const hexPadStart = (data: number | bigint, len: number) => {
  return `0x${data.toString(16).padStart(len, '0')}`
}

const ADDRESS_LEN = 32 * 2 // 32 bytes (hex)

const addresses = (t: TestData) => ({
  bob: t.bob.starknetContract.address,
  alice: t.alice.starknetContract.address,
})

const expectOwner = async (c: StarknetContract, expected: string) => {
  const { response } = await c.call('owner')

  const owner = hexPadStart(response, ADDRESS_LEN)
  expect(owner).to.deep.equal(expected)
}

const expectProposedOwner = async (c: StarknetContract, expected: string) => {
  const { response } = await c.call('proposed_owner')

  const proposedOwner = hexPadStart(response, ADDRESS_LEN)
  expect(proposedOwner).to.deep.equal(expected)
}

// idempotent - end with original owner
export const shouldBehaveLikeOwnableContract = (beforeFn: BeforeFn) => {
  describe('ownable behavior', function () {
    let t: TestData

    before(async () => {
      t = await beforeFn()
    })

    // TODO: test 'owner', 'proposed_owner', 'transfer_ownership', 'accept_ownership' in depth

    it(`should have 'owner' set`, async () => {
      const { alice } = addresses(t)

      // initial owner is alice
      await expectOwner(t.ownable, alice)
    })

    it(`should be able to 'transfer_ownership'`, async () => {
      const { alice, bob } = addresses(t)

      await t.alice.invoke(t.ownable, 'transfer_ownership', {
        new_owner: bob,
      })

      // owner is still alice
      await expectOwner(t.ownable, alice)
      await expectProposedOwner(t.ownable, bob)
    })

    it(`should be able to 'accept_ownership'`, async () => {
      const { bob } = addresses(t)

      await t.bob.invoke(t.ownable, 'accept_ownership')

      // owner is now bob
      await expectOwner(t.ownable, bob)
      await expectProposedOwner(t.ownable, hexPadStart(0, ADDRESS_LEN)) // 0x0
    })

    it(`should be able to 'transfer_ownership' back`, async () => {
      const { alice, bob } = addresses(t)

      await t.bob.invoke(t.ownable, 'transfer_ownership', {
        new_owner: alice,
      })

      // owner is still bob
      await expectOwner(t.ownable, bob)
      await expectProposedOwner(t.ownable, alice)
    })

    it(`should be able to 'accept_ownership' again`, async () => {
      const { alice } = addresses(t)

      await t.alice.invoke(t.ownable, 'accept_ownership')

      // owner is now alice
      await expectOwner(t.ownable, alice)
      await expectProposedOwner(t.ownable, hexPadStart(0, ADDRESS_LEN)) // 0x0
    })

    it(`should fail with account without fees`, async () => {
      const accountNoFees = await starknet.OpenZeppelinAccount.createAccount()

      await t.alice.invoke(t.ownable, 'transfer_ownership', {
        new_owner: accountNoFees.address,
      })

      try {
        await accountNoFees.invoke(t.ownable, 'accept_ownership', { maxFee: 1e18 })
        expect.fail()
      } catch (err: any) {}
    })
  })
}
