import { spawn } from 'child_process'
import { writeFileSync } from 'fs'
import { join } from 'path'
import { CompiledContract } from 'starknet'
import fs from 'fs'

export abstract class StarknetCLICommand {
  protected classHash: string

  constructor() {}

  protected abstract spawnClassHash(): Promise<string>

  public async run(): Promise<string> {
    this.classHash = await this.spawnClassHash()
    return this.classHash
  }
  public rm(contract_path: string) {
    fs.rmSync(contract_path)
  }
}

class CLIClassHashCommand extends StarknetCLICommand {
  private command: string
  private contract: string

  constructor(contract: string) {
    super()
    this.contract = contract
    this.command = 'starknet-class-hash'
  }

  protected async spawnClassHash(): Promise<string> {
    let args = [this.contract]
    let ls = spawn(this.command, args)
    let classHash = ''
    return await new Promise((resolve, reject) => {
      ls.stdout.on('data', (data) => {
        classHash += data.toString()
      })
      ls.on('close', (code) => {
        resolve(classHash)
      })
      ls.on('error', (err) => {
        reject(err)
      })
    })
  }
}

export const starknetClassHash = async (contract: CompiledContract): Promise<string> => {
  const replacer = (key, value) => (typeof value === 'bigint' ? value.toString() : value)
  writeFileSync(join(__dirname, 'contract.json'), JSON.stringify(contract, replacer), {
    flag: 'w',
  })
  const contract_path = join(__dirname, 'contract.json')
  const commandCLI = new CLIClassHashCommand(contract_path)

  const classHash = await commandCLI.run()

  await new Promise((f) => setTimeout(f, 2000))
  commandCLI.rm(contract_path)
  return classHash
}
