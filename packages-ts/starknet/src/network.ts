import { ChildProcess, spawn } from 'child_process'
import axios from 'axios'

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

  private async isServerAlive() {
    try {
      await axios.get(`http://127.0.0.1:5050/is_alive`)
      return true
    } catch (err: unknown) {
      // cannot connect, so address is not occupied
      return false
    }
  }

  public async start(): Promise<void> {
    if (await this.isServerAlive()) {
      this.cleanup()
    }
    this.childProcess = await this.spawnChildProcess()

    // capture the most recent message from stderr
    // Needed to avoid TimeOut error in some of our tests.
    this.childProcess.stderr?.on('data', async () => {})
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

  protected async spawnChildProcess(): Promise<ChildProcess> {
    let args = ['--port', this.port, '--gas-price', '1', '--lite-mode', '--timeout', '30000']
    if (this.opts?.seed) {
      args.push('--seed', this.opts.seed.toString())
    }
    return spawn(this.command, args)
  }

  protected cleanup(): void {
    this.childProcess?.kill()
  }
}

class VenvHardHatNode extends IntegratedDevnet {
  private command: string[]
  private opts: any

  constructor(port: string, opts: any) {
    super(port)

    this.opts = opts
    this.command = ['hardhat', 'node']
  }

  protected async spawnChildProcess(): Promise<ChildProcess> {
    return spawn('npx', this.command)
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

export const startNode = async (opts?: {}): Promise<IntegratedDevnet> => {
  const hardhatNode = new VenvHardHatNode('8545', opts)

  await hardhatNode.start()

  // Starting to poll hardhatNode too soon can result in ENOENT
  await new Promise((f) => setTimeout(f, 2000))

  return hardhatNode
}
