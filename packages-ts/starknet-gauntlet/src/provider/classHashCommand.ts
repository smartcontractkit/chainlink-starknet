import { spawn } from 'child_process'

export abstract class StarknetCLICommand {
  protected classHash: string

  constructor() {}

  protected abstract spawnClassHash(): Promise<string>

  public async run(): Promise<string> {
    this.classHash = await this.spawnClassHash()
    return this.classHash
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

export const starknetClassHash = async (contract: string): Promise<string> => {
  const commandCLI = new CLIClassHashCommand(contract)

  const classHash = await commandCLI.run()

  await new Promise((f) => setTimeout(f, 2000))

  return classHash
}
