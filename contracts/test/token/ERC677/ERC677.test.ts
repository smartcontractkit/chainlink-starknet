import { expect } from 'chai'
import { toBN } from 'starknet/utils/number'
import { starknet } from 'hardhat'
import { uint256 } from 'starknet'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../../constants'
import { account } from '@chainlink/starknet'

describe('ERC677', function () {
  this.timeout(TIMEOUT)
  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let receiverFactory: StarknetContractFactory
  let tokenFactory: StarknetContractFactory
  let receiver: StarknetContract
  let sender: Account
  let token: StarknetContract
  let data: (number | bigint)[]

  beforeEach(async () => {
    sender = await starknet.deployAccount('OpenZeppelin')
    await funder.fund([{ account: sender.address, amount: 5000 }])
    receiverFactory = await starknet.getContractFactory('token677_receiver_mock')
    tokenFactory = await starknet.getContractFactory('link_token')

    receiver = await receiverFactory.deploy({})

    token = await tokenFactory.deploy({ owner: sender.starknetContract.address })

    await sender.invoke(token, 'permissionedMint', {
      account: sender.starknetContract.address,
      amount: uint256.bnToUint256(1000),
    })

    await sender.invoke(token, 'transfer', {
      recipient: sender.starknetContract.address,
      amount: uint256.bnToUint256(100),
    })
    const { value: value } = await receiver.call('getSentValue')
    expect(uint256.uint256ToBN(value)).to.deep.equal(toBN(0))
  })

  describe('#transferAndCall(address, uint, bytes)', () => {
    beforeEach(() => {
      data = [100n, 0n, 12]
    })

    it('transfers the tokens', async () => {
      let { balance: balance } = await token.call('balanceOf', {
        account: receiver.address,
      })

      expect(uint256.uint256ToBN(balance)).to.deep.equal(toBN(0))

      await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: uint256.bnToUint256(100),
        data: data,
      })

      let { balance: balance1 } = await token.call('balanceOf', {
        account: receiver.address,
      })
      expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(100))
    })

    it('calls the token fallback function on transfer', async () => {
      await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: uint256.bnToUint256(100),
        data: data,
      })

      const { bool: bool } = await receiver.call('getCalledFallback', {})
      expect(bool).to.deep.equal(1n)

      const { address: address } = await receiver.call('getTokenSender', {})
      expect(address).to.equal(BigInt(sender.starknetContract.address))

      const { value: sentValue } = await receiver.call('getSentValue')
      expect(uint256.uint256ToBN(sentValue)).to.deep.equal(toBN(100))
    })

    it('transfer succeeds with response', async () => {
      const response = await sender.invoke(token, 'transferAndCall', {
        to: receiver.address,
        value: uint256.bnToUint256(100),
        data: data,
      })
      expect(response).to.exist
    })

    it('throws when the transfer fails', async () => {
      try {
        await sender.invoke(token, 'transfer', {
          recipient: receiver.address,
          amount: uint256.bnToUint256(10000),
        })
        expect.fail()
      } catch (error: any) {}
    })
  })

  describe('when sending to a contract that is not ERC677 compatible', () => {
    let nonERC677: StarknetContract

    beforeEach(async () => {
      const nonERC677Factory = await starknet.getContractFactory('not_erc677_compatible')
      nonERC677 = await nonERC677Factory.deploy({})
      data = [1000n, 0n, 12n]
    })

    it('throws an error', async () => {
      try {
        await sender.invoke(token, 'transferAndCall', {
          to: nonERC677.address,
          value: uint256.bnToUint256(1000),
          data: data,
        })
        expect.fail()
      } catch (error: any) {
        let { balance: balance1 } = await token.call('balanceOf', {
          account: nonERC677.address,
        })
        expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(0))
      }
    })

    it('throws an error when sending to 0 address', async () => {
      try {
        await sender.invoke(token, 'transferAndCall', {
          to: 0,
          value: uint256.bnToUint256(1000),
          data: data,
        })
        expect.fail()
      } catch (error: any) {
        let { balance: balance1 } = await token.call('balanceOf', {
          account: 0,
        })
        expect(uint256.uint256ToBN(balance1)).to.deep.equal(toBN(0))
      }
    })
  })
})
