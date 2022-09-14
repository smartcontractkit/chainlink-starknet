import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployOZCommand from '@chainlink/starknet-gauntlet-oz/src/commands/account/deploy'
import deployTokenCommand from '../../src/commands/token/deploy'
import deployCommand from '../../src/commands/bridge/deploy'
import setL1Bridge from '../../src/commands/bridge/setL1Bridge'
import setL2Token from '../../src/commands/bridge/setL2Token'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContract, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { compressProgram } from 'starknet/dist/utils/stark'

describe('Bridge Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string
  let publicKey: string
  let l1BridgeAddress: number = 42 // mock placeholder
  let bridgeContractAddress: string
  let tokenContractAddress: string

  beforeAll(async () => {
    network = await startNetwork()
  }, 15000)

  it(
    'Deploy OZ Account',
    async () => {
      const command = await registerExecuteCommand(deployOZCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      account = report.responses[0].contract
      privateKey = report.data.privateKey
      publicKey = report.data.publicKey

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
    'Deploy L2 Bridge with Default Wallet as Governor',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          account: account,
          pk: privateKey,
          governor: account,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      bridgeContractAddress = report.responses[0].contract

      const bridge = loadContract(CONTRACT_LIST.BRIDGE)
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

  it(
    'Set L1 Bridge',
    async () => {
      const command = await registerExecuteCommand(setL1Bridge).create(
        {
          account: account,
          pk: privateKey,
          address: l1BridgeAddress,
        },
        [bridgeContractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const bridge = loadContract(CONTRACT_LIST.BRIDGE)
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
          account: account,
          pk: privateKey,
          address: tokenContractAddress,
        },
        [bridgeContractAddress],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const bridge = loadContract(CONTRACT_LIST.BRIDGE)
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
