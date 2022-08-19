import { assert, expect } from 'chai'
import BN from 'bn.js'
import { toBN } from 'starknet/utils/number'
import { starknet } from 'hardhat'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'
import { getSelectorFromName } from 'starknet/dist/utils/hash'

function uint256ToBigInt(uint256: { low: bigint; high: bigint }) {
    return new BN(((BigInt(uint256.high) << BigInt(128)) + BigInt(uint256.low)).toString())
}

describe('LinkToken', function () {
    this.timeout(TIMEOUT)

    let receiverFactory: StarknetContractFactory
    let tokenFactory: StarknetContractFactory

    let receiver: StarknetContract
    let recipient: StarknetContract
    let sender: Account
    let owner: Account
    let token: StarknetContract
    let params: BN[]

    beforeEach(async () => {
        sender = await starknet.deployAccount('OpenZeppelin')
        owner = await starknet.deployAccount('OpenZeppelin')

        receiverFactory = await starknet.getContractFactory('token677_receiver_mock')
        tokenFactory = await starknet.getContractFactory('link_token')

        receiver = await receiverFactory.deploy({})
        token = await tokenFactory.deploy({ minter: owner.starknetContract.address })
        await owner.invoke(token, 'permissionedMint', {
            recipient: owner.starknetContract.address,
            amount: { high: 0n, low: 1000000000000000000000000000n },
        })
    })

    it('assigns all of the balance to the owner', async () => {
        let { balance: balance } = await token.call('balanceOf', {
            account: owner.starknetContract.address,
        })
        expect(uint256ToBigInt(balance).toString()).to.equal('1000000000000000000000000000')
    })

    describe('#transfer(address,uint256)', () => {
        beforeEach(async () => {
            await owner.invoke(token, 'transfer', {
                recipient: sender.starknetContract.address,
                amount: { high: 0n, low: 100n },
            })
            const { value: sentValue } = await receiver.call('get_sent_value')
            expect(uint256ToBigInt(sentValue)).to.deep.equal(toBN(0))
        })

        it('does not let you transfer to the null address', async () => {
            try {
                await sender.invoke(token, 'transfer', { recipient: 0, value: { high: 0n, low: 100n } })
                expect.fail()
            } catch (error: any) {
                let { balance: balance1 } = await token.call('balanceOf', {
                    account: sender.starknetContract.address,
                })
                expect(uint256ToBigInt(balance1)).to.deep.equal(toBN(100))
            }
        })

        // TODO For now it let you transfer to the contract itself
        xit('does not let you transfer to the contract itself', async () => {
            try {
                await sender.invoke(token, 'transfer', { recipient: token.address, amount: { high: 0n, low: 100n } })
                expect.fail()
            } catch (error: any) {
                let { balance: balance1 } = await token.call('balanceOf', {
                    account: sender.starknetContract.address,
                })
                expect(uint256ToBigInt(balance1)).to.deep.equal(toBN(100))
            }
        })

        it('transfers the tokens', async () => {
            let { balance: balance } = await token.call('balanceOf', {
                account: receiver.address,
            })
            expect(uint256ToBigInt(balance)).to.deep.equal(toBN(0))

            await sender.invoke(token, 'transfer', { recipient: receiver.address, amount: { high: 0n, low: 100n } })

            let { balance: balance1 } = await token.call('balanceOf', {
                account: receiver.address,
            })
            expect(uint256ToBigInt(balance1)).to.deep.equal(toBN(100))
        })

        it('does NOT call the fallback on transfer', async () => {
            await sender.invoke(token, 'transfer', { recipient: receiver.address, amount: { high: 0n, low: 100n } })
            const { bool: bool } = await receiver.call('get_called_fallback', {})
            console.log('bool: ', bool)
            expect(bool).to.deep.equal(0n)
        })

        it('transfer succeeds with response', async () => {
            const response = await sender.invoke(token, 'transfer', {
                recipient: receiver.address,
                amount: { high: 0n, low: 100n },
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
            expect(uint256ToBigInt(allowance)).to.deep.equal(toBN(0))

            let { balance: balance } = await token.call('balanceOf', {
                account: recipient.address,
            })
            expect(uint256ToBigInt(balance)).to.deep.equal(toBN(0))
        })

        it('transfers the amount to the contract and calls the contract', async () => {
            let selector = getSelectorFromName('callback_without_withdrawl')
            // data = []
            await owner.invoke(token, 'transferAndCall', {
                to: recipient.address,
                value: { high: 0n, low: 1000n },
                data: [selector],
            })

            let { balance: balance } = await token.call('balanceOf', {
                account: recipient.address,
            })
            expect(uint256ToBigInt(balance)).to.deep.equal(toBN(amount))
            const { remaining: allowance } = await token.call('allowance', {
                owner: owner.starknetContract.address,
                spender: recipient.address,
            })
            expect(uint256ToBigInt(allowance)).to.deep.equal(toBN(0))

            const { bool: fallBack } = await recipient.call('get_fallback', {})
            expect(fallBack).to.deep.equal(1n)

            const { bool: callData } = await recipient.call('get_call_data', {})
            expect(callData).to.deep.equal(1n)
        })
        it('transfers the amount to the contract and calls the contract', async () => {
            let selector = getSelectorFromName('callback_with_withdrawl')
            // data = []
            await owner.invoke(token, 'approve', { spender: recipient.address, amount: { high: 0n, low: 1000n } })

            const { remaining: allowance } = await token.call('allowance', {
                owner: owner.starknetContract.address,
                spender: recipient.address,
            })
            expect(uint256ToBigInt(allowance)).to.deep.equal(toBN(amount))

            await owner.invoke(token, 'transferAndCall', {
                to: recipient.address,
                value: { high: 0n, low: 1000n },
                data: [selector, 0n, 1000n, owner.starknetContract.address, token.address],
            })

            let { balance: balance } = await token.call('balanceOf', {
                account: recipient.address,
            })
            expect(uint256ToBigInt(balance)).to.deep.equal(toBN(amount + amount))

            const { bool: fallBack } = await recipient.call('get_fallback', {})
            expect(fallBack).to.deep.equal(1n)

            const { bool: callData } = await recipient.call('get_call_data', {})
            expect(callData).to.deep.equal(1n)

            const { value: value } = await recipient.call('get_tokens', {})
            expect(uint256ToBigInt(value)).to.deep.equal(toBN(amount))
        })
    })
})
