import { ChildProcess, spawn } from 'child_process'
import net from 'net'
import { DEVNET_NAME, HARDHAT_NAME } from './utils'

export class NetworkManager {
  private opts?: any
  private Devnet: StarknetDevnet
  private Hardhat: HardHatNode

  constructor(opts?: any) {
    this.opts = opts

    if (this.opts.required.length > 2) {
      throw new Error('Too many networks required')
    }

    for (let i = 0; i < this.opts.required.length; i++) {
      const network = this.opts.required[i]
      if (network === 'starknet') {
        if (this.opts.config.starknet === DEVNET_NAME) {
          if (this.Devnet) {
            throw new Error('Starknet network is already up')
          }
          this.Devnet = new StarknetDevnet('5050', this.opts.opts)
        }
      } else if (network === 'ethereum') {
        if (this.opts.config.ethereum === HARDHAT_NAME) {
          if (this.Hardhat) {
            throw new Error('Ethereum network is already up')
          }
          this.Hardhat = new HardHatNode('8545', this.opts.opts)
        }
      } else {
        throw new Error(`Unknown ${network} network`)
      }
    }
  }

  public async start(): Promise<void> {
    if (this.Devnet) {
      await this.Devnet.startNetwork()
    }
    if (this.Hardhat) {
      await this.Hardhat.startNode()
    }
  }

  public stop() {
    if (this.Devnet) {
      this.Devnet.stop()
    }
    if (this.Hardhat) {
      this.Hardhat.stop()
    }
  }

  public async restart(): Promise<void> {
    if (this.Devnet || this.Hardhat) {
      await this.start()
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
  private opts?: string[]

  constructor(port: string, opts?: string[]) {
    super(port)

    this.opts = opts
    this.command = 'starknet-devnet'
  }

  protected async spawnChildProcess(): Promise<ChildProcess> {
    let args = ['--port', this.port, '--gas-price', '1', '--lite-mode', '--timeout', '30000']
    this.opts?.forEach((opt) => {
      args.push(opt)
    })
    return spawn(this.command, args)
  }

  public async startNetwork(): Promise<void> {
    if (!(await this.isFreePort(parseInt(this.port)))) {
      return
    }
    await this.start()

    // Starting to poll devnet too soon can result in ENOENT
    await new Promise((f) => setTimeout(f, 2000))
  }
}

class HardHatNode extends ChildProcessManager {
  private command: string
  private opts?: string[]

  constructor(port: string, opts?: string[]) {
    super(port)

    this.opts = opts
    this.command = 'npx'
    // this.command = ['hardhat', 'node']
    // this.opts.forEach(function (opt) {
    //   this.command.push(opt)
    // })
  }

  protected async spawnChildProcess(): Promise<ChildProcess> {
    let args = ['hardhat', 'node']
    this.opts?.forEach((opt) => {
      args.push(opt)
    })

    return spawn(this.command, args)
  }

  public async startNode(): Promise<void> {
    if (!(await this.isFreePort(parseInt(this.port)))) {
      return
    }
    await this.start()

    // Starting to poll hardhatNode too soon can result in ENOENT
    await new Promise((f) => setTimeout(f, 2000))
  }
}
