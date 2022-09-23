import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployOZCommand from '@chainlink/starknet-gauntlet-oz/src/commands/account/deploy'
import deployOCR2Command from '../../src/commands/ocr2/deploy'
import deployACCommand from '../../src/commands/accessController/deploy'
import deployProxyCommand from '../../src/commands/proxy/deploy'
import proposeAggregatorCommand from '../../src/commands/proxy/proposeAggregator'
import confirmAggregatorCommand from '../../src/commands/proxy/confirmAggregator'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContract_Ocr2, CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract, InvokeTransactionReceiptResponse } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

describe('Proxy Contract', () => {
  let network: IntegratedDevnet
  let account: string
  let privateKey: string
  let contractAddress: string
  let accessController: string
  let proxy: string

  beforeAll(async () => {
    network = await startNetwork()
  }, TIMEOUT)
  
  afterAll(async () => {
    network.stop()
})

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
      const response = await fetch(`${gateway_url}/mint`, {
        method: 'post',
        body: JSON.stringify(body),
        headers: { 'Content-Type': 'application/json' },
      })

      const data = await response.json()
      expect(data.new_balance).toEqual(balance)
    },
    TIMEOUT,
  )

  it(
    'Deploy AC',
    async () => {
      // TODO: owner can't be 0 anymore
      const command = await registerExecuteCommand(deployACCommand).create(
        {
          account: account,
          pk: privateKey,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      accessController = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Deploy OCR2',
    async () => {
      const command = await registerExecuteCommand(deployOCR2Command).create(
        {
          account: account,
          pk: privateKey,
          input: {
            owner: account,
            maxAnswer: 10000,
            minAnswer: 1,
            decimals: 18,
            description: 'Test Feed',
            billingAccessController: accessController,
            linkToken: '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
          },
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
    'Deploy Proxy',
    async () => {
      const command = await registerExecuteCommand(deployProxyCommand).create(
        {
          account: account,
          pk: privateKey,
          input: {
            owner: account,
            address: contractAddress,
          },
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      proxy = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Propose Proxy Aggregator',
    async () => {
      const command = await registerExecuteCommand(proposeAggregatorCommand).create(
        {
          account: account,
          pk: privateKey,
          input: {
            owner: account,
            address: contractAddress,
          },
        },
        [proxy],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  it(
    'Confirm Proxy Aggregator',
    async () => {
      const command = await registerExecuteCommand(confirmAggregatorCommand).create(
        {
          account: account,
          pk: privateKey,
          input: {
            owner: account,
            address: contractAddress,
          },
        },
        [proxy],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )
})
