import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployTokenCommand from '../../src/commands/token/deploy'
import deployL2Bridge from '../../src/commands/L2-bridge/deploy'
import setL1Bridge from '../../src/commands/L2-bridge/setL1Bridge'
import setL2Token from '../../src/commands/L2-bridge/setL2Token'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
  devnetAccount0Address,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadL2Contract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

let account = devnetAccount0Address

describe('Bridge Contract', () => {
  let network: IntegratedDevnet
  let l1BridgeAddress: string = "0xB81C73E1b942C38a3C821605b56A2734Dc31ee12" // random L1 address
  let bridgeContractAddress: string
  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  it(
    'Deploy L2 Bridge with Default Wallet as Governor',
    async () => {
      const command = await registerExecuteCommand(deployL2Bridge).create(
        {
          governor: account,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      bridgeContractAddress = report.responses[0].contract

      const bridge = loadL2Contract(CONTRACT_LIST.L2_BRIDGE)
      const bridgeContract = new Contract(
        bridge.abi,
        bridgeContractAddress,
        makeProvider(LOCAL_URL).provider,
      )
      const response = await bridgeContract.get_governor()
      const governor = response[0]
      expect(governor).toEqual(new BN(account.split('x')[1], 16))
    },
    TIMEOUT,
  )

  it(
    'Deploy Token',
    async () => {
      const command = await registerExecuteCommand(deployTokenCommand).create(
        {
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
    'Set L1 Bridge',
    async () => {
      const command = await registerExecuteCommand(setL1Bridge).create(
        {
          address: l1BridgeAddress,
        },
        [bridgeContractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const bridge = loadL2Contract(CONTRACT_LIST.L2_BRIDGE)
      const bridgeContract = new Contract(
        bridge.abi,
        bridgeContractAddress,
        makeProvider(LOCAL_URL).provider,
      )
      const response = await bridgeContract.get_l1_bridge()

      // TODO: Process response
      console.log(response)
    },
    TIMEOUT,
  )

  it(
    'Set L2 Token',
    async () => {
      const command = await registerExecuteCommand(setL2Token).create(
        {
          address: tokenContractAddress,
        },
        [bridgeContractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const bridge = loadL2Contract(CONTRACT_LIST.L2_BRIDGE)
      const bridgeContract = new Contract(
        bridge.abi,
        bridgeContractAddress,
        makeProvider(LOCAL_URL).provider,
      )
      const response = await bridgeContract.get_l2_token()

      // TODO: Process response
      console.log(response)
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
