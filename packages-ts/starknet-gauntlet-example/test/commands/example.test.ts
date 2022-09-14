import deployOZCommand from '../../../starknet-gauntlet-oz/src/commands/account/deploy'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { makeProvider } from '@chainlink/starknet-gauntlet'
import { Contract } from 'starknet'
import deployCommand from '../../src/commands/example/deploy'
import increaseBalanceCommand from '../../src/commands/example/increaseBalance'
import inspectionCommand from '../../src/commands/example/inspection/inspect'
import {
  LOCAL_URL,
  registerExecuteCommand,
  TIMEOUT,
  registerInspectCommand,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadExampleContract } from '../utils'

const getBalance = async (address: string) => {
  const contract = new Contract(
    loadExampleContract().abi,
    address,
    makeProvider(LOCAL_URL).provider,
  )
  const balance = await contract.get_balance()

  return new BN(balance.res).toString()
}

describe('Example Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string
  let contractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, TIMEOUT)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      account = report.responses[0].contract
      privateKey = report.data.privateKey

      // Fund the newly allocated account
      let gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'
      let balance = 1e21
      const body = {
        address: account,
        amount: balance,
        lite: true,
      }

      try {
        const response = await fetch(`${gateway_url}/mint`, {
          method: 'post',
          body: JSON.stringify(body),
          headers: { 'Content-Type': 'application/json' },
        })

        const data = await response.json()
        expect(data.new_balance).toEqual(balance)
      } catch (e) {
        console.log(e)
      }
    },
    TIMEOUT,
  )

  it(
    'Deployment with default input',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          account: account,
          pk: privateKey,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Increase balance with input flag',
    async () => {
      const command = await registerExecuteCommand(increaseBalanceCommand).create(
        {
          input: { balance: '100' },
          account: account,
          pk: privateKey,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const balance = await getBalance(contractAddress)
      expect(balance).toEqual('100')
    },
    TIMEOUT,
  )

  it(
    'Increase balance with custom flags',
    async () => {
      const command = await registerExecuteCommand(increaseBalanceCommand).create(
        {
          balance: '100',
          account: account,
          pk: privateKey,
        },
        [contractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const balance = await getBalance(contractAddress)
      expect(balance).toEqual('200')
    },
    TIMEOUT,
  )

  it('Inspection of state with no account', async () => {
    const command = await registerInspectCommand(inspectionCommand).create({}, [contractAddress])

    const expectedResult = {
      balance: '200',
    }
    const report = await command.execute()
    expect(report.data.data).toEqual(expectedResult)
  }, 3000)

  // TODO: Include tests with account contract. Pending on devnet to support predeployed accounts and fee token

  afterAll(() => {
    network.stop()
  })
})
