import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account, StarknetContractFactory } from 'hardhat/types/runtime'
import { num, Provider, Contract, hash } from 'starknet'
import { account, expectInvokeError, expectSuccessOrDeclared, expectCallError } from '@chainlink/starknet'
const starkwareCrypto = require('@starkware-industries/starkware-crypto-utils')

describe('upgradeable', function () {
    this.timeout(1_000_000)

    let owner: Account
    let nonOwner: Account
    const opts = account.makeFunderOptsFromEnv()
    const funder = new account.Funder(opts)
    let upgradeableFactory: StarknetContractFactory
    let nonUpgradeableFactory: StarknetContractFactory
    let upgradeableContract: StarknetContract

    const provider = new Provider({ sequencer: { baseUrl: starknet.networkConfig.url! } })

    // should be beforeeach, but that'd be horribly slow. Just remember that the tests are not idempotent
    before(async function () {
        owner = await starknet.OpenZeppelinAccount.createAccount()

        await funder.fund([
            { account: owner.address, amount: 1e21 },
        ])
        await owner.deployAccount()

        upgradeableFactory = await starknet.getContractFactory('mock_upgradeable')
        await expectSuccessOrDeclared(owner.declare(upgradeableFactory, { maxFee: 1e20 }))

        nonUpgradeableFactory = await starknet.getContractFactory('mock_non_upgradeable')
        await expectSuccessOrDeclared(owner.declare(nonUpgradeableFactory, { maxFee: 1e20 }))
    })

    describe('Upgrade contract', () => {

        beforeEach(async () => {
            upgradeableContract = await owner.deploy(upgradeableFactory)
        })

        it('succeeds if class hash exists', async () => {
            expect((await upgradeableContract.call('foo')).response).to.equal(true)

            const newClassHash = await nonUpgradeableFactory.getClassHash()

            await owner.invoke(
                upgradeableContract,
                'upgrade',
                {
                    new_impl: newClassHash
                }
            )

            // should error because the contract has upgraded and no longer has foo function
            await expectCallError(upgradeableContract.call('foo'))

            // we're not able to directly call bar via hardhat StarknetContract because 
            // hardhat is still checking the abi of the old class hash (line below will error)
            // so we use starknet.js directly
            // expect((await upgradeableContract.call('bar')).response).to.equal(true)

            const { abi: testAbi } = await provider.getClassAt(upgradeableContract.address)
            if (testAbi === undefined) { throw new Error("no abi.") }
            const afterUpgradeContract = new Contract(testAbi, upgradeableContract.address, provider)

            expect((await afterUpgradeContract.call('bar')).toString()).to.equal('true')
        })

        it('reverts if new implementation class hash does not exist', async () => {
            expect((await upgradeableContract.call('foo')).response).to.equal(true)

            const nonExistentClassHash = `0x${starkwareCrypto.pedersen(["random", "hash"])}`

            // class hash is not declared
            await expectInvokeError(
                owner.invoke(upgradeableContract, 'upgrade', { new_impl: nonExistentClassHash })
            )

        })
    })
})
