import deployOZCommand from '@chainlink/starknet-gauntlet-oz/src/commands/account/deploy'
import deployTokenCommand from '../../src/commands/token/deploy'
import mintTokensCommand from '../../src/commands/token/mint'
import transferTokensCommand from '../../src/commands/token/transfer'
import balanceOfCommand from '../../src/commands/inspection/balanceOf'
 
import {
  registerExecuteCommand,
  registerInspectCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'

describe('Token Contract', () => {
  let network: IntegratedDevnet

  let defaultAccount: string
  let defaultPk: string
  let defaultBalance: number

  let ozAccount: string
  let ozPk: string
  let ozBalance: number

  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork({ seed: 0 })
    
    // account #0 with seed 0
    defaultAccount = '0x7e00d496e324876bbc8531f2d9a82bf154d1a04a50218ee74cdd372f75a551a'
    defaultPk = '0xe3e70682c2094cac629f6fbed82c07cd'
    defaultBalance = 0
  }, 15000)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      ozAccount = report.responses[0].contract
      ozPk = report.data.privateKey
      ozBalance = 0
    },
    TIMEOUT,
  )

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployTokenCommand).create(
        {
          account: defaultAccount,
          pk: defaultPk,
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
          account: defaultAccount,
          pk: defaultPk,
          recipient: defaultAccount,
          amount,
        },
        [tokenContractAddress],
      )
      let report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      defaultBalance = amount

      const inspectCommand = await registerInspectCommand(balanceOfCommand).create(
        {
          account: defaultAccount,
        },
        [tokenContractAddress]
      )
      report = await inspectCommand.execute()
      expect(report.data?.data?.balance).toEqual(defaultBalance.toString())
    },
    TIMEOUT,
  )

  it(
    'Transfer tokens to OZ account',
    async () => {
      const amount = 50

      const executeCommand = await registerExecuteCommand(transferTokensCommand).create(
        {
          account: defaultAccount,
          pk: defaultPk,
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
          account: ozAccount
        },
        [tokenContractAddress]
      )
      report = await inspectCommand.execute()
      expect(report.data?.data?.balance).toEqual(ozBalance.toString())
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
