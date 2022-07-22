import { makeProvider, makeWallet, Dependencies } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/multisig/deploy'
import setOwners from '../../src/commands/multisig/setOwners'
import setThreshold from '../../src/commands/multisig/setThreshold'
import { wrapCommand } from '../../src/wrapper'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
} from '@chainlink/starknet-gauntlet/test/utils'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'

describe('Multisig', () => {
  let network: IntegratedDevnet
  let multisigContractAddress: string
  const SEED: number = 10
  let accounts: string[] = [
    '0x2ea08d8ea2435755926e3766f9a131c6c0c5b15e690864df4478621565d5d48',
    '0x458d3122886b79be2d4f455ab64e4c8f49fa613829e21a40bb0f61814a7d8e7',
    '0x2095b0b11b236bfd6f337a16d26ff63edf6de7e792523ec90396b906a27b2fa',
  ]
  let publicKeys: string[] = [
    '0x5366dfa9668f9f51c6f4277455b34881262f12cb6b12b487877d9319a5b48bc',
    '0x5ad457b3d822e2f1671c2046038a3bb45e6683895f7a4af266545de03e0d3e9',
    '0x1a9dea7b74c0eee5f1873c43cc600a01ec732183d5b230efa9e945495823e9a',
  ]
  let privateKeys: string[] = [
    '0x7b89296c6dcbac5008577eb1924770d3',
    '0x766bad0734c2da8003cc0f2793fdcab8',
    '0x470b9805d2d6b8777dc59a3ad035d259',
  ]

  let newOwnerAccount = {
    account: '0x53670d100f4d7aca6afca85bb1a7267e494c51331b2eb99f0c0442cfbcc56b1',
    publicKey: '0x6a5f1d67f6b59f3a2a294c3e523731b43fccbb7230985be7399c118498faf03',
    privateKey: '0x8ceac392904cdefcf84b683a749f9c5',
  }

  beforeAll(async () => {
    network = await startNetwork()
  }, 5000)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          owners: accounts,
          threshold: 1,
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      multisigContractAddress = report.responses[0].contract
      console.log(multisigContractAddress)
    },
    TIMEOUT,
  )

  it(
    'Set Threshold with multisig',
    async () => {
      const deps: Dependencies = {
        logger: logger,
        prompt: prompt,
        makeEnv: (flags) => {
          return {
            providerUrl: 'http://127.0.0.1:5050',
            pk: privateKeys[0],
            publicKey: publicKeys[0],
            account: accounts[0],
            multisig: multisigContractAddress,
          }
        },
        makeProvider: makeProvider,
        makeWallet: makeWallet,
      }

      // Create Multisig Proposal
      const command = await wrapCommand(registerExecuteCommand(setThreshold))(deps).create(
        {
          threshold: 2,
        },
        [multisigContractAddress],
      )

      let report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      const multisigProposalId = report.data // TODO: fix this (not sure where msig proposal is in the response object)

      // Approve Multisig Proposal
      const approveCommand = await wrapCommand(registerExecuteCommand(setThreshold))(deps).create(
        {
          threshold: 2,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await approveCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      // Execute Multisig Proposal
      const executeCommand = await wrapCommand(registerExecuteCommand(setThreshold))(deps).create(
        {
          threshold: 2,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  it(
    'Set Owners with multisig',
    async () => {
      const deps: Dependencies = {
        logger: logger,
        prompt: prompt,
        makeEnv: (flags) => {
          return {
            providerUrl: 'http://127.0.0.1:5050',
            pk: privateKeys[0],
            publicKey: publicKeys[0],
            account: accounts[0],
            multisig: multisigContractAddress,
          }
        },
        makeProvider: makeProvider,
        makeWallet: makeWallet,
      }

      accounts.push(newOwnerAccount.account)

      // Create Multisig Proposal
      const command = await wrapCommand(registerExecuteCommand(setOwners))(deps).create(
        {
          owners: accounts,
        },
        [multisigContractAddress],
      )

      let report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      const multisigProposalId = report.data // TODO: fix this (not sure where msig proposal is in the response object)

      // Approve Multisig Proposal
      const approveCommand = await wrapCommand(registerExecuteCommand(setOwners))(deps).create(
        {
          owners: accounts,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await approveCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      // Execute Multisig Proposal
      const executeCommand = await wrapCommand(registerExecuteCommand(setOwners))(deps).create(
        {
          owners: accounts,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
