import { assert, expect } from 'chai'
import BN from 'bn.js'
import { toBN } from 'starknet/utils/number'
import { starknet } from 'hardhat'
import { constants, ec, encode, hash, number, uint256, stark, KeyPair } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'

function uint256ToBigInt(uint256: { low: bigint; high: bigint }) {
  return new BN(((BigInt(uint256.high) << BigInt(128)) + BigInt(uint256.low)).toString())
}

describe('ERC677', function () {
  this.timeout(TIMEOUT)

  let receiverFactory: StarknetContractFactory
  let tokenFactory: StarknetContractFactory

  let receiver: StarknetContract
  let sender: Account
  let token: StarknetContract
  let data: (number | bigint)[]

  beforeEach(async () => {
    sender = await starknet.deployAccount('OpenZeppelin')

    receiverFactory = await starknet.getContractFactory('token677ReceiverMock')
    tokenFactory = await starknet.getContractFactory('token677')

    receiver = await receiverFactory.deploy({})
    token = await tokenFactory.deploy({ minter: sender.starknetContract.address })

    await sender.invoke(token, 'permissionedMint', {
      recipient: sender.starknetContract.address,
      amount: { high: 0n, low: 1000n },
    })
    await sender.invoke(token, 'transfer', {
      recipient: sender.starknetContract.address,
      amount: { high: 0n, low: 100n },
    })
    const { value: value } = await receiver.call('get_sent_value')
    expect(uint256ToBigInt(value)).to.deep.equal(toBN(0))
  })

  describe('#transferAndCall(address, uint, bytes)', () => {
    beforeEach(() => {
      data = [100n, 0n, 12]
    })

    it('transfers the tokens', async () => {
      let { balance: balance } = await token.call('balanceOf', {
        account: receiver.address,
      })

      expect(uint256ToBigInt(balance)).to.deep.equal(toBN(0))

      await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: { high: 0n, low: 100n },
        selector: 0,
        data: data,
      })

      let { balance: balance1 } = await token.call('balanceOf', {
        account: receiver.address,
      })
      expect(uint256ToBigInt(balance1)).to.deep.equal(toBN(100))
    })

    it('calls the token fallback function on transfer', async () => {
      await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: { high: 0n, low: 100n },
        selector: 0,
        data: data,
      })

      const { bool: bool } = await receiver.call('get_called_fallback', {})
      expect(bool).to.deep.equal(1n)

      const { address: address } = await receiver.call('get_token_sender', {})
      expect(address).to.equal(BigInt(sender.starknetContract.address))

      const { value: sentValue } = await receiver.call('get_sent_value')
      expect(uint256ToBigInt(sentValue)).to.deep.equal(toBN(100))
    })

    it('transfer succeeds with response', async () => {
      const response = await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: { high: 0n, low: 100n },
        selector: 0,
        data: data,
      })
      expect(response).to.exist
    })

    it('throws when the transfer fails', async () => {
      try {
        await sender.invoke(token, 'transfer', { recipient: receiver.address, amount: { high: 0n, low: 10000n } })
        expect.fail()
      } catch (error: any) {}
    })
  })

  describe('when sending to a contract that is not ERC677 compatible', () => {
    let nonERC677: StarknetContract

    beforeEach(async () => {
      const nonERC677Factory = await starknet.getContractFactory('notERC677Compatible')
      nonERC677 = await nonERC677Factory.deploy({})
      data = [1000n, 0n, 12n]
    })

    it('throws an error', async () => {
      try {
        await sender.invoke(token, 'transferAndCall', {
          to: nonERC677.address,
          value: { high: 0n, low: 1000n },
          selector: 0,
          data: data,
        })
        expect.fail()
      } catch (error: any) {
        let { balance: balance1 } = await token.call('balanceOf', {
          account: nonERC677.address,
        })
        expect(uint256ToBigInt(balance1)).to.deep.equal(toBN(0))
      }
    })
  })
})
