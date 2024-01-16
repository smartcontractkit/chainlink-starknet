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
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadExampleContract } from '../utils'

const getBalance = async (address: string) => {
  const contract = new Contract(
    loadExampleContract().abi,
    address,
    makeProvider(LOCAL_URL).provider,
  )
  const balance = await contract.get_balance()

  return BigInt(balance.res).toString()
}

describe('Example Contract', () => {
  let contractAddress: string

  it(
    'Deployment with default input',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

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
})
