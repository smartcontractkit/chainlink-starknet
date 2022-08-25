import { expect } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { uint256 } from 'starknet'
import { toBN } from 'starknet/utils/number'
import { TIMEOUT } from '../../../constants'

export type TestData = {
  token: StarknetContract
  owner: Account
  alice: Account
  bob: Account
}

export type BeforeFn = () => Promise<TestData>

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
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(15))
      }

      await t.owner.invoke(t.token, 'permissionedMint', {
        recipient: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(12),
      })
      {
        const { totalSupply: totalSupply } = await t.token.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(27))

        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(12))
      }
    })

    it('should burn successfully', async () => {
      /* Burn some token with the good minter and check the user's balance */
      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      {
        const { totalSupply: totalSupply } = await t.token.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(24))

        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(12))
      }

      await t.owner.invoke(t.token, 'permissionedBurn', {
        account: t.bob.starknetContract.address,
        amount: uint256.bnToUint256(10),
      })
      {
        const { totalSupply: totalSupply } = await t.token.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(14))

        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(2))
      }
    })

    it('should burn fail because amount bigger than balance', async () => {
      /* Burn some token with the good minter but with an amount bigger than the balance */
      /* All test should fail */
      try {
        await t.owner.invoke(t.token, 'permissionedBurn', {
          account: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}

      try {
        await t.owner.invoke(t.token, 'permissionedBurn', {
          account: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })

    it('should burn fail because wrong minter', async () => {
      /* Burn some token with the wrong minter */
      /* All test should fail */
      try {
        await t.alice.invoke(t.token, 'permissionedBurn', {
          account: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
      }

      try {
        await t.bob.invoke(t.token, 'permissionedBurn', {
          account: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        throw new Error('This should not pass!')
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
      {
        await t.alice.invoke(t.token, 'transfer', {
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(3),
        })
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(9))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance1)
        expect(sum_balance1).to.deep.equal(toBN(8))
      }

      await t.bob.invoke(t.token, 'transfer', {
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(4))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(13))
      }
    })

    it('should transfer fail because amount bigger than balance', async () => {
      /* Transfer some token from one user to another with amout bigger than balance */
      /* All tests should fail */
      try {
        await t.bob.invoke(t.token, 'transfer', {
          recipient: t.alice.starknetContract.address,
          amount: uint256.bnToUint256(12),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await t.alice.invoke(t.token, 'transfer', {
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(17),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
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
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(10))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(7))
      }

      await t.owner.invoke(t.token, 'transferFrom', {
        sender: t.bob.starknetContract.address,
        recipient: t.alice.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(14))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(3))
      }
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
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.alice.starknetContract.address,
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(300),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
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
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(11))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(6))
      }
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
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(0))

        const { balance: balance1 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(17))
      }
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
      {
        const { balance: balance } = await t.token.call('balanceOf', {
          account: t.alice.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(1))

        const { balance: balance2 } = await t.token.call('balanceOf', {
          account: t.bob.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance2)
        expect(sum_balance2).to.deep.equal(toBN(16))
      }
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
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await t.owner.invoke(t.token, 'transferFrom', {
          sender: t.alice.starknetContract.address,
          recipient: t.bob.starknetContract.address,
          amount: uint256.bnToUint256(208),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })
  })
}
