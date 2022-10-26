import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { hexPadStart } from '../../utils'

export type BeforeFn = () => Promise<TestData>
export type TestData = {
  ownable: StarknetContract
  alice: Account
  bob: Account
}

const ADDRESS_LEN = 32 * 2 // 32 bytes (hex)

const addresses = (t: TestData) => ({
  bob: t.bob.starknetContract.address,
  alice: t.alice.starknetContract.address,
})

const expectOwner = async (c: StarknetContract, expected: string) => {
  const { owner: raw } = await c.call('owner')

  const owner = hexPadStart(raw, ADDRESS_LEN)
  expect(owner).to.deep.equal(expected)
}

const expectProposedOwner = async (c: StarknetContract, expected: string) => {
  const { proposed_owner: raw } = await c.call('proposed_owner')

  const proposedOwner = hexPadStart(raw, ADDRESS_LEN)
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
      const accountNoFees = await starknet.deployAccount('OpenZeppelin')

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
