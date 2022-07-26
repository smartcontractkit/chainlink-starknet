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

  constructor(port: string) {
    super(port)

    this.command = 'starknet-devnet'
  }

  protected async spawnChildProcess(): Promise<ChildProcess> {
    return spawn(this.command, ['--port', this.port, '--gas-price', '1'])
  }

  protected cleanup(): void {
    this.childProcess?.kill()
  }
}

export const startNetwork = async (): Promise<IntegratedDevnet> => {
  const devnet = new VenvDevnet('5050')

  await devnet.start()

  return devnet
}
