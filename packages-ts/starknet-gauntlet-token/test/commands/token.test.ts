import deployOZCommand from '@chainlink/starknet-gauntlet-oz/src/commands/account/deploy'
import deployTokenCommand from '../../src/commands/token/deploy'
import mintTokensCommand from '../../src/commands/token/mint'
import transferTokensCommand from '../../src/commands/token/transfer'
import balanceOfCommand from '../../src/commands/inspection/balanceOf'
import {
  StarknetAccount,
  fetchAccount,
  registerExecuteCommand,
  registerInspectCommand,
  TIMEOUT,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Token Contract', () => {
  let defaultAccount: StarknetAccount

  let ozAccount: string
  let ozBalance: number

  let tokenContractAddress: string

  beforeAll(async () => {
    // account #0 with seed 0
    defaultAccount = await fetchAccount()
  }, TIMEOUT)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      ozAccount = report.responses[0].contract
      ozBalance = 0
    },
    TIMEOUT,
  )

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployTokenCommand).create(
        {
          account: defaultAccount.address,
          pk: defaultAccount.privateKey,
          link: true,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      tokenContractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Mint tokens for Default account',
    async () => {
      const amount = 10000000

      const executeCommand = await registerExecuteCommand(mintTokensCommand).create(
        {
          account: defaultAccount.address,
          pk: defaultAccount.privateKey,
          recipient: defaultAccount.address,
          amount,
        },
        [tokenContractAddress],
      )
      let report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      defaultAccount.balance = amount

      const inspectCommand = await registerInspectCommand(balanceOfCommand).create(
        {
          address: defaultAccount.address,
        },
        [tokenContractAddress],
      )
      report = await inspectCommand.execute()
      expect(report.data?.data?.balance).toEqual(defaultAccount.balance.toString())
    },
    TIMEOUT,
  )

  it(
    'Transfer tokens to OZ account',
    async () => {
      const amount = 50

      const executeCommand = await registerExecuteCommand(transferTokensCommand).create(
        {
          account: defaultAccount.address,
          pk: defaultAccount.privateKey,
          recipient: ozAccount,
          amount,
        },
        [tokenContractAddress],
      )
      let report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      ozBalance = amount

      const inspectCommand = await registerInspectCommand(balanceOfCommand).create(
        {
          address: ozAccount,
        },
        [tokenContractAddress],
      )
      report = await inspectCommand.execute()
      expect(report.data?.data?.balance).toEqual(ozBalance.toString())
    },
    TIMEOUT,
  )
})
