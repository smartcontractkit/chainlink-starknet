import { ChildProcess, spawn } from 'child_process'

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

  protected async spawnChildProcess(): Promise<ChildProcess> {
    let args = ['--port', this.port, '--gas-price', '1', '--lite-mode']
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
