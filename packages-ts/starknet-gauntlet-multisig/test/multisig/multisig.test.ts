import { makeProvider, makeWallet, Dependencies } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/multisig/deploy'
import setSigners from '../../src/commands/multisig/setSigners'
import setThreshold from '../../src/commands/multisig/setThreshold'
import { wrapCommand } from '../../src/wrapper'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/starknet-gauntlet/test/utils'
import { logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import { loadContract } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract } from 'starknet'

describe('Multisig', () => {
  let multisigContractAddress: string
  const SEED: number = 10
  const accounts: string[] = [
    '0x78662e7352d062084b0010068b99288486c2d8b914f6e2a55ce945f8792c8b1',
    '0x49dfb8ce986e21d354ac93ea65e6a11f639c1934ea253e5ff14ca62eca0f38e',
    '0x4f348398f859a55a0c80b1446c5fdc37edb3a8478a32f10764659fc241027d3',
  ]
  const publicKeys: string[] = [
    '0x7a1bb2744a7dd29bffd44341dbd78008adb4bc11733601e7eddff322ada9cb',
    '0xb8fd4ddd415902d96f61b7ad201022d495997c2dff8eb9e0eb86253e30fabc',
    '0x5e05d2510c6110bde03df9c1c126a1f592207d78cd9e481ac98540d5336d23c',
  ]
  const privateKeys: string[] = [
    '0xe1406455b7d66b1690803be066cbe5e',
    '0xa20a02f0ac53692d144b20cb371a60d7',
    '0xa641611c17d4d92bd0790074e34beeb7',
  ]

  const newSignerAccount = {
    account: '0xd513de92c16aa42418cf7e5b60f8022dbee1b4dfd81bcf03ebee079cfb5cb5',
    publicKey: '0x4708e28e2424381659ea6b7dded2b3aff4b99debfcf6080160a9d098ac2214d',
    privateKey: '0x5b4ac23628a5749277bcabbf4726b025',
  }

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
      const myFlags = {
        providerUrl: LOCAL_URL,
        pk: privateKeys[0],
        publicKey: publicKeys[0],
        account: accounts[0],
        multisig: multisigContractAddress,
      }
      const deps: Dependencies = {
        logger: logger,
        prompt: prompt,
        makeEnv: (flags) => {
          return myFlags
        },
        makeProvider: makeProvider,
        makeWallet: makeWallet,
      }

      // Create Multisig Proposal
      const command = await wrapCommand(registerExecuteCommand(setThreshold))(deps).create(
        {
          threshold: 2,
          ...myFlags
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
          ...myFlags
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
          ...myFlags
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
      const myFlags = (index) => {
        return {
          providerUrl: LOCAL_URL,
          pk: privateKeys[index],
          publicKey: publicKeys[index],
          account: accounts[index],
          multisig: multisigContractAddress,
        }
      }
      const deps = (index: number): Dependencies => {
        return {
          logger: logger,
          prompt: prompt,
          makeEnv: (flags) => {
            return myFlags(index)
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
          ...myFlags(0)
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
          ...myFlags(0)
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
          ...myFlags(1)
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
          ...myFlags(0)
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
})
