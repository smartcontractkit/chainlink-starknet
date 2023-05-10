import { expect } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { uint256, num } from 'starknet'
import { TIMEOUT } from '../../../constants'
import { expectInvokeError } from '@chainlink/starknet/src/utils'

export type BeforeFn = () => Promise<TestData>
export type TestData = {
  token: StarknetContract
  owner: Account
  alice: Account
  bob: Account
}

const addresses = (t: TestData) => ({
  owner: t.owner.starknetContract.address,
  bob: t.bob.starknetContract.address,
  alice: t.alice.starknetContract.address,
})

const expectERC20Balance = async (token: StarknetContract, acc: Account, expected: number) => {
  const { balance: raw } = await token.call('balanceOf', {
    account: acc.starknetContract.address,
  })
  const balance = uint256.uint256ToBN(raw)
  expect(balance).to.deep.equal(num.toBigInt(expected))
}

const expectERC20TotalSupply = async (token: StarknetContract, expected: number) => {
  const { totalSupply: raw } = await token.call('totalSupply', {})
  const totalSupply = uint256.uint256ToBN(raw)
  expect(totalSupply).to.deep.equal(num.toBigInt(expected))
}

export const shouldBehaveLikeStarkGateERC20 = (beforeFn: BeforeFn) => {
  describe('StarkGate.ERC20 behavior', function () {
    this.timeout(TIMEOUT)

    let t: TestData

    before(async () => {
      t = await beforeFn()
    })

    it(`should 'permissionedMint' successfully (2x)`, async () => {
      const { alice, bob } = addresses(t)

      await t.owner.invoke(t.token, 'permissionedMint', {
        account: alice,
        amount: uint256.bnToUint256(15),
      })

      await expectERC20Balance(t.token, t.alice, 15)

      await t.owner.invoke(t.token, 'permissionedMint', {
        account: bob,
        amount: uint256.bnToUint256(12),
      })

      await expectERC20TotalSupply(t.token, 27)
      await expectERC20Balance(t.token, t.bob, 12)
    })

    it(`should 'permissionedBurn' successfully (2x)`, async () => {
      const { alice, bob } = addresses(t)

      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: alice,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20TotalSupply(t.token, 24)
      await expectERC20Balance(t.token, t.alice, 12)

      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: bob,
        amount: uint256.bnToUint256(10),
      })

      await expectERC20TotalSupply(t.token, 14)
      await expectERC20Balance(t.token, t.bob, 2)
    })

    it(`reverts on 'permissionedBurn' (amount > balance)`, async () => {
      const { alice, bob } = addresses(t)

      await expectInvokeError(
        t.owner.invoke(t.token, 'permissionedBurn', {
          account: bob,
          amount: uint256.bnToUint256(103),
        }),
      )

      await expectInvokeError(
        t.owner.invoke(t.token, 'permissionedBurn', {
          account: alice,
          amount: uint256.bnToUint256(189),
        }),
      )
    })

    it(`reverts on 'permissionedBurn' without permission`, async () => {
      const { alice, bob } = addresses(t)

      await expectInvokeError(
        t.alice.invoke(t.token, 'permissionedBurn', {
          account: bob,
          amount: uint256.bnToUint256(103),
        }),
      )

      await expectInvokeError(
        t.bob.invoke(t.token, 'permissionedBurn', {
          account: alice,
          amount: uint256.bnToUint256(189),
        }),
      )
    })

    it(`should 'permissionedMint' and 'transfer' successfully`, async () => {
      const { alice, bob } = addresses(t)

      await t.owner.invoke(t.token, 'permissionedMint', {
        account: bob,
        amount: uint256.bnToUint256(3),
      })
      await t.alice.invoke(t.token, 'transfer', {
        recipient: bob,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20Balance(t.token, t.alice, 9)
      await expectERC20Balance(t.token, t.bob, 8)

      await t.bob.invoke(t.token, 'transfer', {
        recipient: alice,
        amount: uint256.bnToUint256(4),
      })

      await expectERC20Balance(t.token, t.alice, 13)
      await expectERC20Balance(t.token, t.bob, 4)
    })

    it(`reverts on 'transfer' (amount > balance)`, async () => {
      const { alice, bob } = addresses(t)

      await expectInvokeError(
        t.bob.invoke(t.token, 'transfer', {
          recipient: alice,
          amount: uint256.bnToUint256(12),
        }),
        'ERC20: transfer amount exceeds balance',
      )

      await expectInvokeError(
        t.alice.invoke(t.token, 'transfer', {
          recipient: bob,
          amount: uint256.bnToUint256(17),
        }),
        'ERC20: transfer amount exceeds balance',
      )
    })

    it(`should 'increaseAllowance' and 'transferFrom' successfully - #1`, async () => {
      const { owner, alice, bob } = addresses(t)

      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(7),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(7),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: alice,
        recipient: bob,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20Balance(t.token, t.alice, 10)
      await expectERC20Balance(t.token, t.bob, 7)

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: bob,
        recipient: alice,
        amount: uint256.bnToUint256(4),
      })

      await expectERC20Balance(t.token, t.alice, 14)
      await expectERC20Balance(t.token, t.bob, 3)
    })

    it(`should 'increaseAllowance' and 'transferFrom' successfully - #2`, async () => {
      const { owner, alice, bob } = addresses(t)

      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(7),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(7),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: alice,
        recipient: bob,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20Balance(t.token, t.alice, 11)
      await expectERC20Balance(t.token, t.bob, 6)

      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(15),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: owner,
        added_value: uint256.bnToUint256(15),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: alice,
        recipient: bob,
        amount: uint256.bnToUint256(11),
      })

      await expectERC20Balance(t.token, t.alice, 0)
      await expectERC20Balance(t.token, t.bob, 17)
    })

    it(`should 'decreaseAllowance' and 'transferFrom' successfully`, async () => {
      const { owner, alice, bob } = addresses(t)

      await t.bob.invoke(t.token, 'decreaseAllowance', {
        spender: owner,
        subtracted_value: uint256.bnToUint256(10),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: bob,
        recipient: alice,
        amount: uint256.bnToUint256(1),
      })

      await expectERC20Balance(t.token, t.alice, 1)
      await expectERC20Balance(t.token, t.bob, 16)
    })

    it(`reverts on 'transferFrom' (amount > allowance)`, async () => {
      const { alice, bob } = addresses(t)

      await expectInvokeError(
        t.owner.invoke(t.token, 'transferFrom', {
          sender: bob,
          recipient: alice,
          amount: { low: 8n, high: 10n },
        }),
        'ERC20: insufficient allowance',
      )

      await expectInvokeError(
        t.owner.invoke(t.token, 'transferFrom', {
          sender: alice,
          recipient: bob,
          amount: uint256.bnToUint256(208),
        }),
        'ERC20: insufficient allowance',
      )
    })
  })
}
