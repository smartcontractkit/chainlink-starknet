import { ChildProcess, spawn } from 'child_process'
import fs from 'fs'
import path from 'path'

export abstract class IntegratedDevnet {
  protected childProcess: ChildProcess

  constructor(protected port: string) {
    IntegratedDevnet.cleanupFns.push(this.cleanup.bind(this))
  }

  protected static cleanupFns: Array<() => void> = []

  public static cleanAll(): void {
    this.cleanupFns.forEach((fn) => fn())
  }

  protected abstract spawnChildProcess(): Promise<ChildProcess>

  protected abstract cleanup(): void

  public async start(): Promise<void> {
    this.childProcess = await this.spawnChildProcess()

    return new Promise((resolve, reject) => {
      setTimeout(resolve, 4000)

      this.childProcess.on('error', (error) => {
        reject(error)
      })
    })
  }

  public stop() {
    if (!this.childProcess) {
      return
    }

    this.cleanup()
  }
}

class VenvDevnet extends IntegratedDevnet {
  private command: string
  private opts: any

  constructor(port: string, opts: any) {
    super(port)

    this.opts = opts
    this.command = 'starknet-devnet'
  }

  protected spawnChildProcess(): Promise<ChildProcess> {
    return new Promise((resolve, reject) => {
      const args = ['--port', this.port, '--gas-price', '1']
      if (this.opts?.seed) {
        args.push('--seed', this.opts.seed.toString())
      } else {
        args.push('--seed', '0')
      }
      console.log('Spawning starknet-devnet:', args.join(' '))
      const childProcess = spawn(this.command, args)
      childProcess.on('error', reject)

      // starknet-devnet takes time to run the starknet rust compiler once first to get the version.
      // This calls `cargo run` and requires building on first run, which can take a while. Wait for some program output before we resolve.
      // ref: https://github.com/0xSpaceShard/starknet-devnet/blob/b7388321471e504a04c831dbc175d5a569b76f0c/starknet_devnet/devnet_config.py#L214
      childProcess.stdout.setEncoding('utf-8')
      let initialOutput = ''
      childProcess.stdout.on('data', (chunk) => {
        initialOutput += chunk
        if (initialOutput.indexOf('listening on') >= 0) {
          console.log('Started starknet-devnet')
          childProcess.stdout.removeAllListeners('data')
          resolve(childProcess)
        }
      })
    })
  }

  protected cleanup(): void {
    this.childProcess?.kill()
  }
}

export const startNetwork = async (opts?: {}): Promise<IntegratedDevnet> => {
  const devnet = new VenvDevnet('5050', opts)

  await devnet.start()

  // Starting to poll devnet too soon can result in ENOENT
  await new Promise((f) => setTimeout(f, 2000))

  return devnet
}
