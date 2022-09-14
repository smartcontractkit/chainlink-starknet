import {
  executeCommands as OCR2ExecuteCommands,
  inspectionCommands as OCR2InspectionCommands,
} from '@chainlink/starknet-gauntlet-ocr2'
import {
  executeCommands as ExampleExecuteCommands,
  inspectionCommands as ExampleInspectionsCommands,
} from '@chainlink/starknet-gauntlet-example'
import { Commands as OZCommands } from '@chainlink/starknet-gauntlet-oz'
import {
  L1Commands as L1StarkgateCommands,
  L2Commands as L2StarkgateCommands,
} from '@chainlink/starknet-gauntlet-starkgate'
import { Commands as ArgentCommands } from '@chainlink/starknet-gauntlet-argent'
import {
  executeCommands as MultisigExecuteCommands,
  inspectionCommands as MultisigInspectionCommands,
  wrapCommand as multisigWrapCommand,
} from '@chainlink/starknet-gauntlet-multisig'

import { executeCLI } from '@chainlink/gauntlet-core'
import { existsSync } from 'fs'
import path from 'path'
import { io, logger, prompt } from '@chainlink/gauntlet-core/dist/utils'
import {
  CommandCtor,
  Dependencies,
  Env,
  ExecuteCommandInstance,
  InspectCommandInstance,
  makeProvider,
  makeWallet,
} from '@chainlink/starknet-gauntlet'
import {
  EVMExecuteCommandInstance,
  CommandCtor as EVMCommandCtor,
  makeWallet as EVMMakeWallet,
  makeProvider as EVMMakeProvider,
  EVMDependencies,
} from '@chainlink/evm-gauntlet'

export const noopPrompt: typeof prompt = async () => {}

const registerExecuteCommand = <UI, CI>(
  registerCommand: (deps: Dependencies) => CommandCtor<ExecuteCommandInstance<UI, CI>>,
  emptyPrompt = false,
) => {
  const deps: Dependencies | Omit<Dependencies, 'makeWallet'> = {
    logger: logger,
    prompt: emptyPrompt ? noopPrompt : prompt,
    makeEnv: (flags) => {
      const env: Env = {
        providerUrl: process.env.NODE_URL || 'https://alpha4.starknet.io',
        pk: process.env.PRIVATE_KEY,
        publicKey: process.env.PUBLIC_KEY,
        account: process.env.ACCOUNT,
        multisig: process.env.MULTISIG,
        billingAccessController: process.env.BILLING_ACCESS_CONTROLLER,
        link: process.env.LINK,
        secret: flags.secret || process.env.SECRET,
      }
      return env
    },
    makeProvider: makeProvider,
    makeWallet: makeWallet,
  }
  return registerCommand(deps)
}

const registerEVMExecuteCommand = <UI, CI>(
  registerCommand: (deps: EVMDependencies) => EVMCommandCtor<EVMExecuteCommandInstance<UI, CI>>,
  gauntletConfig,
) => {
  const deps: EVMDependencies = {
    logger: logger,
    prompt: prompt,
    makeEnv: (flags) => {
      return {
        providerUrl:
          process.env.NODE_URL || 'https://goerli.infura.io/v3/7c43471f9d604276a856f0cff1edb645',
        pk: process.env.PRIVATE_KEY,
      }
    },
    makeProvider: EVMMakeProvider,
    makeWallet: EVMMakeWallet,
  }
  return registerCommand(deps)
}

const registerInspectionCommand = <QueryResult>(
  registerCommand: (
    deps: Omit<Dependencies, 'makeWallet'>,
  ) => CommandCtor<InspectCommandInstance<QueryResult>>,
) => {
  const deps: Omit<Dependencies, 'makeWallet'> = {
    logger: logger,
    prompt: prompt,
    makeEnv: (flags) => {
      const env: Env = {
        providerUrl: process.env.NODE_URL || 'https://alpha4.starknet.io',
      }
      return env
    },
    makeProvider: makeProvider,
  }
  return registerCommand(deps)
}

const L1ExecuteCommands = [...L1StarkgateCommands]
const L2ExecuteCommands = [
  ...OCR2ExecuteCommands,
  ...ExampleExecuteCommands,
  ...OZCommands,
  ...L2StarkgateCommands,
  ...ArgentCommands,
  ...MultisigExecuteCommands,
]

const msigCommands = L2ExecuteCommands.map((c) => registerExecuteCommand(c, true)).map(
  multisigWrapCommand,
)
const unregistedInspectionCommands = [
  ...ExampleInspectionsCommands,
  ...MultisigInspectionCommands,
  ...OCR2InspectionCommands,
]

const commands = {
  custom: [
    ...L2ExecuteCommands.map((c) => registerExecuteCommand(c)),
    ...L1ExecuteCommands.map((c) => registerEVMExecuteCommand(c, null)),
    ...msigCommands.map((c) => registerExecuteCommand(c)),
    ...unregistedInspectionCommands.map(registerInspectionCommand),
  ],
  loadDefaultFlags: () => ({}),
  abstract: {
    findPolymorphic: () => undefined,
    makeCommand: () => undefined,
  },
}

;(async () => {
  try {
    const networkPossiblePaths = [
      path.join(process.cwd(), 'networks'),
      path.join(__dirname, '../networks'),
    ]
    const networkPath = networkPossiblePaths.filter((networkPath) => existsSync(networkPath))[0]
    const result = await executeCLI(commands, networkPath)
    if (result) {
      io.saveJSON(result, process.env['REPORT_NAME'] ? process.env['REPORT_NAME'] : 'report')
    }
    process.exit(0)
  } catch (e) {
    console.log(e)
    console.log('Starknet Command execution error', e.message)
    process.exitCode = 1
  }
})()
