import { STARKNET_DEVNET_URL } from './constants'
import { execSync } from 'node:child_process'
import { Account } from 'starknet'
import * as path from 'node:path'
import { ethers } from 'hardhat'
import { json } from 'starknet'
import * as fs from 'node:fs'

export type FetchStarknetAccountParams = Readonly<{
  accountIndex?: number
}>

export const fetchStarknetAccount = async (params?: FetchStarknetAccountParams) => {
  const response = await fetch(`${STARKNET_DEVNET_URL}/predeployed_accounts`)
  const accounts = await response.json()
  const accIndex = params?.accountIndex ?? 0

  const account = accounts.at(accIndex)
  if (account == null) {
    throw new Error(`no account available at index ${accIndex}`)
  }

  return new Account(
    {
      nodeUrl: STARKNET_DEVNET_URL,
    },
    account.address,
    account.private_key,
  )
}

export const getStarknetContractArtifacts = (name: string) => {
  const rootDir = getRootDir()
  return {
    contract: getStarknetContractArtifactPath(rootDir, name, false),
    casm: getStarknetContractArtifactPath(rootDir, name, true),
  }
}

export const waitForTransaction = async (
  tx: () => ReturnType<typeof ethers.provider.sendTransaction>,
) => {
  const result = await tx()
  return await result.wait()
}

export const waitForTransactions = async (
  txs: (() => ReturnType<typeof ethers.provider.sendTransaction>)[],
  cb?: () => void | Promise<void>,
) => {
  const results = new Array<Awaited<ReturnType<typeof waitForTransaction>>>()
  for (const tx of txs) {
    results.push(await waitForTransaction(tx))
    cb != null && (await cb())
  }
  return results
}

const getRootDir = () => {
  const result = execSync('git rev-parse --show-toplevel').toString()
  return result.replace(/\n/g, '')
}

const getStarknetContractArtifactPath = (rootDir: string, name: string, casm: boolean) => {
  return json.parse(
    fs
      .readFileSync(
        path.join(
          rootDir,
          'contracts',
          'target',
          'release',
          `chainlink_${name}.${casm ? 'compiled_' : ''}contract_class.json`,
        ),
      )
      .toString('ascii'),
  )
}
