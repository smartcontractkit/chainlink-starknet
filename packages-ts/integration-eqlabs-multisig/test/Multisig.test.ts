import { expect } from 'chai'
import { starknet } from 'hardhat'
import { StarknetContract, Account } from 'hardhat/types/runtime'
import { getSelectorFromName } from 'starknet/dist/utils/hash'
import { number } from 'starknet'
import { account, loadConfig, NetworkManager, FunderOptions, Funder } from '@chainlink/starknet'

describe('Multisig integration tests', function () {
  this.timeout(300_000)

  const config = loadConfig()
  const optsConf = { config, required: ['starknet'] }
  const manager = new NetworkManager(optsConf)

  let opts: FunderOptions
  let funder: Funder

  let account1: Account
  let account2: Account
  let account3: Account

  let multisig: StarknetContract

  before(async function () {
    await manager.start()
    opts = account.makeFunderOptsFromEnv()
    funder = new account.Funder(opts)

    account1 = await starknet.deployAccount('OpenZeppelin')
    account2 = await starknet.deployAccount('OpenZeppelin')
    account3 = await starknet.deployAccount('OpenZeppelin')

    await funder.fund([
      { account: account1.address, amount: 5000 },
      { account: account2.address, amount: 5000 },
      { account: account3.address, amount: 5000 },
    ])
  })

  it('Deploy contract', async () => {
    let multisigFactory = await starknet.getContractFactory('Multisig')
    multisig = await multisigFactory.deploy({
      signers: [
        number.toBN(account1.starknetContract.address),
        number.toBN(account2.starknetContract.address),
        number.toBN(account3.starknetContract.address),
      ],
      threshold: 2,
    })

    expect(multisig).to.be.ok
  })

  it('should submit & confirm transaction', async () => {
    const nonce = 0
    const newThreshold = 1n
    const selector = getSelectorFromName('set_threshold')

    const payload = {
      to: multisig.address,
      function_selector: selector,
      calldata: [newThreshold],
      nonce,
    }

    {
      const res = await account1.invoke(multisig, 'submit_transaction', payload)
      const txReciept = await starknet.getTransactionReceipt(res)

      expect(txReciept.events.length).to.equal(1)
      expect(txReciept.events[0].data.length).to.equal(3)
      expect(txReciept.events[0].data[1]).to.equal(number.toHex(number.toBN(nonce, 'hex')))
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
      expect(res.threshold).to.equal(newThreshold)
    }
  })

  after(async function () {
    manager.stop()
  })
})
