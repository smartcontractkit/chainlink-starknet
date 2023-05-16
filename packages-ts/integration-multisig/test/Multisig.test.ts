import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { num, hash } from 'starknet'
import { account, expectSuccessOrDeclared } from '@chainlink/starknet'

describe('Multisig integration tests', function () {
  this.timeout(300_000)

  const opts = account.makeFunderOptsFromEnv()
  const funder = new account.Funder(opts)

  let account1: Account
  let account2: Account
  let account3: Account

  let multisig: StarknetContract

  before(async function () {
    account1 = await starknet.OpenZeppelinAccount.createAccount()
    account2 = await starknet.OpenZeppelinAccount.createAccount()
    account3 = await starknet.OpenZeppelinAccount.createAccount()

    await funder.fund([
      { account: account1.address, amount: 1e21 },
      { account: account2.address, amount: 1e21 },
      { account: account3.address, amount: 1e21 },
    ])
    await account1.deployAccount()
    await account2.deployAccount()
    await account3.deployAccount()
  })

  it('Deploy contract', async () => {
    let multisigFactory = await starknet.getContractFactory('multisig')
    await expectSuccessOrDeclared(account1.declare(multisigFactory, { maxFee: 1e20 }))

    multisig = await account1.deploy(multisigFactory, {
      signers: [
        num.toBigInt(account1.starknetContract.address),
        num.toBigInt(account2.starknetContract.address),
        num.toBigInt(account3.starknetContract.address),
      ],
      threshold: 2,
    })

    expect(multisig).to.be.ok
  })

  it('should submit & confirm transaction', async () => {
    const nonce = 0
    const newThreshold = 1n
    const selector = hash.getSelectorFromName('set_threshold')

    const payload = {
      to: multisig.address,
      function_selector: selector,
      calldata: [newThreshold],
      nonce,
    }

    {
      const res = await account1.invoke(multisig, 'submit_transaction', payload)
      const txReciept = await starknet.getTransactionReceipt(res)

      expect(txReciept.events.length).to.equal(2)
      expect(txReciept.events[0].data.length).to.equal(3)
      expect(txReciept.events[0].data[1]).to.equal(num.toHex(num.toBigInt(nonce, 'hex')))
    }

    await account1.invoke(multisig, 'confirm_transaction', {
      nonce,
    })

    await account2.invoke(multisig, 'confirm_transaction', {
      nonce,
    })

    await account3.invoke(multisig, 'execute_transaction', {
      nonce,
    })

    {
      const res = await multisig.call('get_threshold')
      expect(res.response).to.equal(newThreshold)
    }
  })
})
