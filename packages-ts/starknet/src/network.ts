import { ChildProcess, spawn } from 'child_process'
import net from 'net'
import { DEVNET_NAME, HARDHAT_NAME } from './utils'

export class NetworkManager {
  private opts?: any
  private strategyDevnet: StarknetDevnet
  private strategyHardhat: HardHatNode

  constructor(opts?: any) {
    this.opts = opts
    if (this.opts?.required[0] === 'starknet') {
      if (this.opts.config.starknet === DEVNET_NAME) {
        this.strategyDevnet = new StarknetDevnet('5050', this.opts)
      }
    }
    if (this.opts?.required[1] === 'ethereum') {
      if (this.opts.config.ethereum === HARDHAT_NAME) {
        this.strategyHardhat = new HardHatNode('8545', this.opts)
      }
    }
  }

  public async start(): Promise<void> {
    if (this.strategyDevnet) {
      await this.strategyDevnet.startNetwork()
    }
    if (this.strategyHardhat) {
      await this.strategyHardhat.startNode()
    }
  }

  public stop() {
    if (this.strategyDevnet) {
      this.strategyDevnet.stop()
    }
    if (this.strategyHardhat) {
      this.strategyHardhat.stop()
    }
  }
}

abstract class ChildProcessManager {
  protected childProcess: ChildProcess

  constructor(protected port: string) {
    ChildProcessManager.cleanupFns.push(this.cleanup.bind(this))
  }

  protected static cleanupFns: Array<() => void> = []

  public static cleanAll(): void {
    this.cleanupFns.forEach((fn) => fn())
  }

  protected abstract spawnChildProcess(): Promise<ChildProcess>

  public async isFreePort(port: number): Promise<boolean> {
    return await new Promise((accept, reject) => {
      const sock = net.createConnection(port)
      sock.once('connect', () => {
        sock.end()
        accept(false)
      })
      sock.once('error', (e: NodeJS.ErrnoException) => {
        sock.destroy()
        if (e.code === 'ECONNREFUSED') {
          accept(true)
        } else {
          reject(e)
        }
      })
    })
  }

  public async start(): Promise<void> {
    this.childProcess = await this.spawnChildProcess()

    // capture the most recent message from stderr
    // Needed to avoid TimeOut error in some of our tests.
    this.childProcess.stderr?.on('data', async () => {})
    return new Promise((resolve, reject) => {
      setTimeout(resolve, 4000)

      this.childProcess.on('error', (error) => {
        console.log('ERROR: ', error)
        reject(error)
      })
    })
  }

  public cleanup(): void {
    this.childProcess?.kill()
  }

  public stop() {
    if (!this.childProcess) {
      return
    }

    this.cleanup()
  }
}

class StarknetDevnet extends ChildProcessManager {
  private command: string
  private opts?: any

  constructor(port: string, opts?: any) {
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

  public async startNetwork(): Promise<void> {
    if (!(await this.isFreePort(parseInt(this.port)))) {
      return
    }
    await this.start()

    // Starting to poll devnet too soon can result in ENOENT
    await new Promise((f) => setTimeout(f, 4000))
  }
}

class HardHatNode extends ChildProcessManager {
  private command: string[]
  private opts?: any

  constructor(port: string, opts?: any) {
    super(port)

    this.opts = opts
    this.command = ['hardhat', 'node']
  }

  protected async spawnChildProcess(): Promise<ChildProcess> {
    return spawn('npx', this.command)
  }

  public async startNode(): Promise<void> {
    if (!(await this.isFreePort(parseInt(this.port)))) {
      return
    }
    await this.start()

    // Starting to poll hardhatNode too soon can result in ENOENT
    await new Promise((f) => setTimeout(f, 10000))
  }
}
