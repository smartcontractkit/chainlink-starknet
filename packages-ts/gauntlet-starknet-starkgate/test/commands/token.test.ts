import { startNetwork, IntegratedDevnet, makeProvider } from '@chainlink/gauntlet-starknet'
import deployOZCommand from '../../../gauntlet-starknet-oz/src/commands/account/deploy'
import deployCommand from '../../src/commands/token/deploy'
import mintCommand from '../../src/commands/token/mint'
import transferCommand from '../../src/commands/token/transfer'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/gauntlet-starknet-example/test/utils'
import { loadContract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { compressProgram } from 'starknet/dist/utils/stark'

describe('Token Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string
  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      account = report.responses[0].contract
      privateKey = report.data.privateKey
    },
    TIMEOUT,
  )

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          link: true,
          account: account,
          pk: privateKey,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      tokenContractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  // TODO: Test mint

  // TODO: Test transfer

  afterAll(() => {
    network.stop()
  })
})
