import { expect } from 'chai'
import { toBN } from 'starknet/utils/number'
import { starknet } from 'hardhat'
import { uint256 } from 'starknet'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'
import { getSelectorFromName } from 'starknet/dist/utils/hash'

describe('LinkToken', function () {
  this.timeout(TIMEOUT)

  let receiverFactory: StarknetContractFactory
  let tokenFactory: StarknetContractFactory
  let receiver: StarknetContract
  let recipient: StarknetContract
  let sender: Account
  let owner: Account
  let token: StarknetContract

  beforeEach(async () => {
    sender = await starknet.deployAccount('OpenZeppelin')
    owner = await starknet.deployAccount('OpenZeppelin')

    receiverFactory = await starknet.getContractFactory('token677_receiver_mock')
    tokenFactory = await starknet.getContractFactory('link_token')

    receiver = await receiverFactory.deploy({})
    token = await tokenFactory.deploy({})
    await owner.invoke(token, 'link_initializer', {
      initial_supply: uint256.bnToUint256(1000000000000000),
      recipient: owner.starknetContract.address,
      proxy_admin: owner.starknetContract.address,
    })
  })

  it('assigns all of the balance to the owner', async () => {
    let { balance: balance } = await token.call('balanceOf', {
      account: owner.starknetContract.address,
    })
    expect(uint256.uint256ToBN(balance).toString()).to.equal('1000000000000000')
  })

  describe('#transfer(address,uint256)', () => {
    beforeEach(async () => {
      await owner.invoke(token, 'transfer', {
        recipient: sender.starknetContract.address,
        amount: uint256.bnToUint256(100),
      })
      const { value: sentValue } = await receiver.call('getSentValue')
      expect(uint256.uint256ToBN(sentValue)).to.deep.equal(toBN(0))
    })

    it('does not let you transfer to the null address', async () => {
      try {
        await sender.invoke(token, 'transfer', { recipient: 0, value: uint256.bnToUint256(100) })
        expect.fail()
      } catch (error: any) {
        let { balance: balance1 } = await token.call('balanceOf', {
          account: sender.starknetContract.address,
        })
        expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(100))
      }
    })

    // TODO For now it let you transfer to the contract itself
    xit('does not let you transfer to the contract itself', async () => {
      try {
        await sender.invoke(token, 'transfer', { recipient: token.address, amount: uint256.bnToUint256(100) })
        expect.fail()
      } catch (error: any) {
        let { balance: balance1 } = await token.call('balanceOf', {
          account: sender.starknetContract.address,
        })
        expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(100))
      }
    })

    it('transfers the tokens', async () => {
      let { balance: balance } = await token.call('balanceOf', {
        account: receiver.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(toBN(0))

      await sender.invoke(token, 'transfer', { recipient: receiver.address, amount: uint256.bnToUint256(100) })

      let { balance: balance1 } = await token.call('balanceOf', {
        account: receiver.address,
      })
      expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(100))
    })

    it('does NOT call the fallback on transfer', async () => {
      await sender.invoke(token, 'transfer', { recipient: receiver.address, amount: uint256.bnToUint256(100) })
      const { bool: bool } = await receiver.call('getCalledFallback', {})
      expect(bool).to.deep.equal(0n)
    })

    it('transfer succeeds with response', async () => {
      const response = await sender.invoke(token, 'transfer', {
        recipient: receiver.address,
        amount: uint256.bnToUint256(100),
      })
      expect(response).to.exist
    })
  })

  describe('#transferAndCall(address,uint256,bytes)', () => {
    const amount = 1000

    before(async () => {
      const receiverFactory = await starknet.getContractFactory('link_receiver')
      const classHash = await receiverFactory.declare()
      recipient = await receiverFactory.deploy({ class_hash: classHash })
      const { remaining: allowance } = await token.call('allowance', {
        owner: owner.starknetContract.address,
        spender: recipient.address,
      })
      expect(uint256.uint256ToBN(allowance)).to.deep.equal(toBN(0))

      let { balance: balance } = await token.call('balanceOf', {
        account: recipient.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(toBN(0))
    })

    it('transfers the amount to the contract and calls the contract function without withdrawl', async () => {
      let selector = getSelectorFromName('callbackWithoutWithdrawl')
      await owner.invoke(token, 'transferAndCall', {
        to: recipient.address,
        value: uint256.bnToUint256(1000),
        data: [selector],
      })

      let { balance: balance } = await token.call('balanceOf', {
        account: recipient.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(toBN(amount))
      const { remaining: allowance } = await token.call('allowance', {
        owner: owner.starknetContract.address,
        spender: recipient.address,
      })
      expect(uint256.uint256ToBN(allowance)).to.deep.equal(toBN(0))

      const { bool: fallBack } = await recipient.call('getFallback', {})
      expect(fallBack).to.deep.equal(1n)

      const { bool: callData } = await recipient.call('getCallData', {})
      expect(callData).to.deep.equal(1n)
    })

    it('transfers the amount to the contract and calls the contract function with withdrawl', async () => {
      let selector = getSelectorFromName('callbackWithWithdrawl')
      await owner.invoke(token, 'approve', { spender: recipient.address, amount: uint256.bnToUint256(1000) })

      const { remaining: allowance } = await token.call('allowance', {
        owner: owner.starknetContract.address,
        spender: recipient.address,
      })
      expect(uint256.uint256ToBN(allowance)).to.deep.equal(toBN(amount))

      await owner.invoke(token, 'transferAndCall', {
        to: recipient.address,
        value: uint256.bnToUint256(1000),
        data: [selector, 0n, 1000n, owner.starknetContract.address, token.address],
      })

      let { balance: balance } = await token.call('balanceOf', {
        account: recipient.address,
      })
      expect(uint256.uint256ToBN(balance)).to.deep.equal(toBN(amount + amount))

      const { bool: fallBack } = await recipient.call('getFallback', {})
      expect(fallBack).to.deep.equal(1n)

      const { bool: callData } = await recipient.call('getCallData', {})
      expect(callData).to.deep.equal(1n)

      const { value: value } = await recipient.call('getTokens', {})
      expect(uint256.uint256ToBN(value)).to.deep.equal(toBN(amount))
    })

    it('transfers the amount to the account and does not call the contract', async () => {
      await owner.invoke(token, 'approve', {
        spender: sender.starknetContract.address,
        amount: uint256.bnToUint256(1000),
      })

      const { remaining: allowance } = await token.call('allowance', {
        owner: owner.starknetContract.address,
        spender: sender.starknetContract.address,
      })
      expect(uint256.uint256ToBN(allowance)).to.deep.equal(toBN(amount))

      await owner.invoke(token, 'transferAndCall', {
        to: sender.starknetContract.address,
        value: uint256.bnToUint256(1000),
        data: [],
      })

      let { balance: balance2 } = await token.call('balanceOf', {
        account: sender.starknetContract.address,
      })
      expect(uint256.uint256ToBN(balance2)).to.deep.equal(toBN(amount))
    })
  })
})
