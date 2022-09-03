import { expect } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { uint256 } from 'starknet'
import { toBN } from 'starknet/utils/number'
import { TIMEOUT } from '../../../constants'
import { assertErrorMsg } from '../../../../test/utils'

export type TestData = {
  token: StarknetContract
  owner: Account
  alice: Account
  bob: Account
}

export type BeforeFn = () => Promise<TestData>

const expectERC20Balance = async (token: StarknetContract, acc: Account, expected: number) => {
  const { balance: raw } = await token.call('balanceOf', {
    account: acc.starknetContract.address,
  })
  const balance = uint256.uint256ToBN(raw)
  expect(balance).to.deep.equal(toBN(expected))
}

const expectERC20TotalSupply = async (token: StarknetContract, expected: number) => {
  const { totalSupply: raw } = await token.call('totalSupply', {})
  const totalSupply = uint256.uint256ToBN(raw)
  expect(totalSupply).to.deep.equal(toBN(expected))
}

export const shouldBehaveLikeStarkGateERC20 = (beforeFn: BeforeFn) => {
  describe('ContractERC20Token', function () {
    this.timeout(TIMEOUT)

    let t: TestData

    before(async () => {
      t = await beforeFn()
    })

    it('should mint successfully', async () => {
      /* Mint some token with the good minter and check the user's balance */
      await t.owner.invoke(t.token, 'permissionedMint', {
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(15),
      })

      await expectERC20Balance(t.token, t.alice, 15)

      await t.owner.invoke(t.token, 'permissionedMint', {
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(12),
      })

      await expectERC20TotalSupply(t.token, 27)
      await expectERC20Balance(t.token, t.bob, 12)
    })

    it('should burn successfully', async () => {
      /* Burn some token with the good minter and check the user's balance */
      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20TotalSupply(t.token, 24)
      await expectERC20Balance(t.token, t.alice, 12)

      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(10),
      })

      await expectERC20TotalSupply(t.token, 14)
      await expectERC20Balance(t.token, t.bob, 2)
    })

    it('should burn fail because amount bigger than balance', async () => {
      /* Burn some token with the good minter but with an amount bigger than the balance */
      /* All test should fail */
      try {
        await t.owner.invoke(t.token, 'permissionedBurn', {
          account: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }

      try {
        await t.owner.invoke(t.token, 'permissionedBurn', {
          account: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
    })

    it('should burn fail because wrong minter', async () => {
      /* Burn some token with the wrong minter */
      /* All test should fail */
      try {
        await t.alice.invoke(t.token, 'permissionedBurn', {
          account: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
      }

      try {
        await t.bob.invoke(t.token, 'permissionedBurn', {
          account: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
      }
    })

    it('should transfer successfully', async () => {
      /* Transfer some token from one user to another and check balance of both users */
      await t.owner.invoke(t.token, 'permissionedMint', {
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      await t.alice.invoke(t.token, 'transfer', {
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20Balance(t.token, t.alice, 9)
      await expectERC20Balance(t.token, t.bob, 8)

      await t.bob.invoke(t.token, 'transfer', {
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })

      await expectERC20Balance(t.token, t.alice, 13)
      await expectERC20Balance(t.token, t.bob, 4)
    })

    it('should transfer fail because amount bigger than balance', async () => {
      /* Transfer some token from one user to another with amout bigger than balance */
      /* All tests should fail */
      try {
        await t.bob.invoke(t.token, 'transfer', {
          recipient: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(12),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
      try {
        await t.alice.invoke(t.token, 'transfer', {
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(17),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
    })

    it('should transferFrom successfully', async () => {
      /* Increase balance then use transferFrom to transfer some token from one user to another and check balance of both users */

      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.alice.starknetContract.address,
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })


      await expectERC20Balance(t.token, t.alice, 10)
      await expectERC20Balance(t.token, t.bob, 7)

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.bob.starknetContract.address,
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })

      await expectERC20Balance(t.token, t.alice, 14)
      await expectERC20Balance(t.token, t.bob, 3)
    })

    it('should transferFrom fail because amount bigger than allowance', async () => {
      /* Use transferFrom to transfer some token from one user to another */
      /* All test should fail because amount bigger than allowance */
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.bob.starknetContract.address,
          recipient: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(200),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.alice.starknetContract.address,
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(300),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
    })

    it('should increase alllowance and transfer some tokens successfully', async () => {
      /* Increase allowance and check if we can use transferFrom to transfer some tokens */
      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.alice.starknetContract.address,
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })

      await expectERC20Balance(t.token, t.alice, 11)
      await expectERC20Balance(t.token, t.bob, 6)

      await t.bob.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(15),
      })
      await t.alice.invoke(t.token, 'increaseAllowance', {
        spender: t.owner.starknetContract.address,
        added_value: uint256.bnToUint256(15),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.alice.starknetContract.address,
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(11),
      })

      await expectERC20Balance(t.token, t.alice, 0)
      await expectERC20Balance(t.token, t.bob, 17)
    })

    it('should decrease alllowance and transfer successfully', async () => {
      /* Decrease allowance and check if we can use transferFrom to transfer some tokens */
      await t.bob.invoke(t.token, 'decreaseAllowance', {
        spender: t.owner.starknetContract.address,
        subtracted_value: uint256.bnToUint256(10),
      })

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.bob.starknetContract.address,
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(1),
      })

      await expectERC20Balance(t.token, t.alice, 1)
      await expectERC20Balance(t.token, t.bob, 16)
    })

    it('should transferFrom fail because amount bigger than allowance', async () => {
      /* Increase allowance and check if we can use transferFrom to transfer some tokens */
      /* All test should fail because amount bigger than allowance */
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.bob.starknetContract.address,
          recipient: t.alice.starknetContract.address,
          amount: { low: 8n, high: 10n },
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.alice.starknetContract.address,
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(208),
        })
        expect.fail()
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
        assertErrorMsg(error?.message, 'SafeUint256: subtraction overflow')
      }
    })
  })
}
