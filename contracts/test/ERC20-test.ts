import { starknet } from 'hardhat'
import { expect } from 'chai'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { toBN } from 'starknet/utils/number'
import { TIMEOUT } from './constants'
import { uint256 } from 'starknet'

export const shouldBehaveLikeERC20Upgradeable = (name: string) => {
  describe('ContractERC20Token', function () {
    this.timeout(TIMEOUT)
    let accountMinter: Account
    let accountUser1: Account
    let accountUser2: Account
    let ERC20Contract: StarknetContract

    before(async () => {
      accountMinter = await starknet.deployAccount('OpenZeppelin')

      let ERC20Factory = await starknet.getContractFactory(name)
      ERC20Contract = await ERC20Factory.deploy({})

      accountUser1 = await starknet.deployAccount('OpenZeppelin')

      accountUser2 = await starknet.deployAccount('OpenZeppelin')

      await accountMinter.invoke(ERC20Contract, 'initializer', {
        initial_supply: uint256.bnToUint256(0),
        recipient: accountMinter.starknetContract.address,
        proxy_admin: accountMinter.starknetContract.address,
      })
    })

    it('should mint successfully', async () => {
      /* Mint some token with the good minter and check the user's balance */
      await accountMinter.invoke(ERC20Contract, 'permissionedMint', {
        recipient: accountUser1.starknetContract.address,
        amount: uint256.bnToUint256(15),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(15))
      }

      await accountMinter.invoke(ERC20Contract, 'permissionedMint', {
        recipient: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(12),
      })
      {
        const { totalSupply: totalSupply } = await ERC20Contract.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(27))

        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(12))
      }
    })

    it('should burn successfully', async () => {
      /* Burn some token with the good minter and check the user's balance */
      await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {
        account: accountUser1.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      {
        const { totalSupply: totalSupply } = await ERC20Contract.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(24))

        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(12))
      }

      await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {
        account: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(10),
      })
      {
        const { totalSupply: totalSupply } = await ERC20Contract.call('totalSupply', {})
        let supply = uint256.uint256ToBN(totalSupply)
        expect(supply).to.deep.equal(toBN(14))

        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(2))
      }
    })

    it('should burn fail because amount bigger than balance', async () => {
      /* Burn some token with the good minter but with an amount bigger than the balance */
      /* All test should fail */
      try {
        await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {
          account: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}

      try {
        await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {
          account: accountUser1.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })

    it('should burn fail because wrong minter', async () => {
      /* Burn some token with the wrong minter */
      /* All test should fail */
      try {
        await accountUser1.invoke(ERC20Contract, 'permissionedBurn', {
          account: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(103),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
      }

      try {
        await accountUser2.invoke(ERC20Contract, 'permissionedBurn', {
          account: accountUser1.starknetContract.address,
          amount: uint256.bnToUint256(189),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {
        expect(/assert/gi.test(error.message)).to.be.true
      }
    })

    it('should transfer successfully', async () => {
      /* Transfer some token from one user to another and check balance of both users */
      await accountMinter.invoke(ERC20Contract, 'permissionedMint', {
        recipient: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      {
        await accountUser1.invoke(ERC20Contract, 'transfer', {
          recipient: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(3),
        })
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance = uint256.uint256ToBN(balance)
        expect(sum_balance).to.deep.equal(toBN(9))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance1)
        expect(sum_balance1).to.deep.equal(toBN(8))
      }

      await accountUser2.invoke(ERC20Contract, 'transfer', {
        recipient: accountUser1.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(4))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(13))
      }
    })

    it('should transfer fail because amount bigger than balance', async () => {
      /* Transfer some token from one user to another with amout bigger than balance */
      /* All tests should fail */
      try {
        await accountUser2.invoke(ERC20Contract, 'transfer', {
          recipient: accountUser1.starknetContract.address,
          amount: uint256.bnToUint256(12),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await accountUser1.invoke(ERC20Contract, 'transfer', {
          recipient: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(17),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })

    it('should transferFrom successfully', async () => {
      /* Increase balance then use transferFrom to transfer some token from one user to another and check balance of both users */

      await accountUser2.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })
      await accountUser1.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })

      await accountMinter.invoke(ERC20Contract, 'transferFrom', {
        sender: accountUser1.starknetContract.address,
        recipient: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(10))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(7))
      }

      await accountMinter.invoke(ERC20Contract, 'transferFrom', {
        sender: accountUser2.starknetContract.address,
        recipient: accountUser1.starknetContract.address,
        amount: uint256.bnToUint256(4),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(14))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(3))
      }
    })

    it('should transferFrom fail because amount bigger than allowance', async () => {
      /* Use transferFrom to transfer some token from one user to another */
      /* All test should fail because amount bigger than allowance */
      try {
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {
          sender: accountUser2.starknetContract.address,
          recipient: accountUser1.starknetContract.address,
          amount: uint256.bnToUint256(200),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {
          sender: accountUser1.starknetContract.address,
          recipient: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(300),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })

    it('should increase alllowance and transfer some tokens successfully', async () => {
      /* Increase allowance and check if we can use transferFrom to transfer some tokens */
      await accountUser2.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })
      await accountUser1.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(7),
      })

      await accountMinter.invoke(ERC20Contract, 'transferFrom', {
        sender: accountUser1.starknetContract.address,
        recipient: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(3),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(11))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(6))
      }
      await accountUser2.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(15),
      })
      await accountUser1.invoke(ERC20Contract, 'increaseAllowance', {
        spender: accountMinter.starknetContract.address,
        added_value: uint256.bnToUint256(15),
      })

      await accountMinter.invoke(ERC20Contract, 'transferFrom', {
        sender: accountUser1.starknetContract.address,
        recipient: accountUser2.starknetContract.address,
        amount: uint256.bnToUint256(11),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(0))

        const { balance: balance1 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance1)
        expect(sum_balance2).to.deep.equal(toBN(17))
      }
    })

    it('should decrease alllowance and transfer successfully', async () => {
      /* Decrease allowance and check if we can use transferFrom to transfer some tokens */
      await accountUser2.invoke(ERC20Contract, 'decreaseAllowance', {
        spender: accountMinter.starknetContract.address,
        subtracted_value: uint256.bnToUint256(10),
      })

      await accountMinter.invoke(ERC20Contract, 'transferFrom', {
        sender: accountUser2.starknetContract.address,
        recipient: accountUser1.starknetContract.address,
        amount: uint256.bnToUint256(1),
      })
      {
        const { balance: balance } = await ERC20Contract.call('balanceOf', {
          account: accountUser1.starknetContract.address,
        })
        let sum_balance1 = uint256.uint256ToBN(balance)
        expect(sum_balance1).to.deep.equal(toBN(1))

        const { balance: balance2 } = await ERC20Contract.call('balanceOf', {
          account: accountUser2.starknetContract.address,
        })
        let sum_balance2 = uint256.uint256ToBN(balance2)
        expect(sum_balance2).to.deep.equal(toBN(16))
      }
    })

    it('should transferFrom fail because amount bigger than allowance', async () => {
      /* Increase allowance and check if we can use transferFrom to transfer some tokens */
      /* All test should fail because amount bigger than allowance */
      try {
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {
          sender: accountUser2.starknetContract.address,
          recipient: accountUser1.starknetContract.address,
          amount: { low: 8n, high: 10n },
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
      try {
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {
          sender: accountUser1.starknetContract.address,
          recipient: accountUser2.starknetContract.address,
          amount: uint256.bnToUint256(208),
        })
        throw new Error('This should not pass!')
      } catch (error: any) {}
    })
  })
}
