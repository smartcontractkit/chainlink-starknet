import { makeProvider, makeWallet, Dependencies } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/multisig/deploy'
import setSigners from '../../src/commands/multisig/setSigners'
import setThreshold from '../../src/commands/multisig/setThreshold'
import { wrapCommand } from '../../src/wrapper'
import {
  registerExecuteCommand,
  TIMEOUT,
  startNetwork,
  IntegratedDevnet,
  LOCAL_URL,
} from '@chainlink/starknet-gauntlet/test/utils'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { loadContract } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'

describe('Multisig', () => {
  let network: IntegratedDevnet
  let multisigContractAddress: string
  const SEED: number = 10
  let accounts: string[] = [
    '0x72ad5b6e5a1c114c370eeabbe700cac4fdd7be47f2ada87ad3b89d346303dec',
    '0x467e04fd5fee4f8a0023c5da41dcdfd963f7e4a9f498a29cbc5c8d030924f5d',
    '0x77c3cd3e09a70580db713fc0d8da7298a9739dea23d4de561eac2991cb6c300',
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

  let newSignerAccount = {
    account: '0x5cdb30a922a2d4f9836877ed76c67564ec32625458884d0f1f2aef1ae023249',
    publicKey: '0x6a5f1d67f6b59f3a2a294c3e523731b43fccbb7230985be7399c118498faf03',
    privateKey: '0x8ceac392904cdefcf84b683a749f9c5',
  }

  beforeAll(async () => {
    network = await startNetwork({ seed: SEED })
  }, TIMEOUT)

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          signers: accounts,
          threshold: 1,
          pk: privateKeys[0],
          publicKey: publicKeys[0],
          account: accounts[0],
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
            providerUrl: LOCAL_URL,
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
      const multisigProposalId = report.data.proposalId

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

      const { contract } = loadContract(CONTRACT_LIST.MULTISIG)
      const multisigContract = new Contract(
        contract.abi,
        multisigContractAddress,
        makeProvider(LOCAL_URL).provider,
      )
      const threshold = await multisigContract.get_threshold()
      expect(Number(threshold)).toEqual(2)
    },
    TIMEOUT,
  )

  it(
    'Set Signers with multisig',
    async () => {
      let deps = (index: number): Dependencies => {
        return {
          logger: logger,
          prompt: prompt,
          makeEnv: (flags) => {
            return {
              providerUrl: LOCAL_URL,
              pk: privateKeys[index],
              publicKey: publicKeys[index],
              account: accounts[index],
              multisig: multisigContractAddress,
            }
          },
          makeProvider: makeProvider,
          makeWallet: makeWallet,
        }
      }

      accounts.push(newSignerAccount.account)

      // Create Multisig Proposal
      const command = await wrapCommand(registerExecuteCommand(setSigners))(deps(0)).create(
        {
          signers: accounts,
        },
        [multisigContractAddress],
      )

      let report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      const multisigProposalId = report.data.proposalId

      // Approve Multisig Proposal (with account 0)
      let approveCommand = await wrapCommand(registerExecuteCommand(setSigners))(deps(0)).create(
        {
          signers: accounts,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await approveCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      // Approve Multisig Proposal (with account 1)
      approveCommand = await wrapCommand(registerExecuteCommand(setSigners))(deps(1)).create(
        {
          signers: accounts,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await approveCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      // Execute Multisig Proposal
      const executeCommand = await wrapCommand(registerExecuteCommand(setSigners))(deps(0)).create(
        {
          signers: accounts,
          multisigProposal: multisigProposalId,
        },
        [multisigContractAddress],
      )

      report = await executeCommand.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      const { contract } = loadContract(CONTRACT_LIST.MULTISIG)
      const multisigContract = new Contract(
        contract.abi,
        multisigContractAddress,
        makeProvider(LOCAL_URL).provider,
      )
      const signers = await multisigContract.get_signers()
      expect(signers).toHaveLength(4)
    },
    TIMEOUT,
  )

  afterAll(() => {
    network.stop()
  })
})
