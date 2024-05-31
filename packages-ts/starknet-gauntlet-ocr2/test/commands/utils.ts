import * as fsp from 'fs/promises'
import * as fs from 'fs'
import path from 'path'

export interface BillingConfig {
  gasBase: string
  gasPerSignature: string
  observationPaymentGjuels: string
  transmissionPaymentGjuels: string
}

export class RDDTempFile {
  private readonly contracts = new Map<
    string,
    {
      billing: BillingConfig
    }
  >()

  constructor(public readonly filepath: string) {
    if (path.extname(filepath) !== '.json') {
      throw new Error(`filepath must point to a json file: ${filepath}`)
    }
  }

  getConfig() {
    return {
      contracts: Object.fromEntries(this.contracts.entries()),
    }
  }

  setBilling(addr: string, data: BillingConfig) {
    const contract = this.contracts.get(addr)
    if (contract == null) {
      this.contracts.set(addr, { billing: data })
    } else {
      this.contracts.set(addr, { ...contract, billing: data })
    }
  }

  async writeFile() {
    await fsp.mkdir(path.dirname(this.filepath), { recursive: true })
    return await fsp.writeFile(this.filepath, JSON.stringify(this.getConfig(), null, 2))
  }

  async removeFile() {
    if (fs.existsSync(this.filepath)) {
      return await fsp.rm(this.filepath)
    }
  }
}
