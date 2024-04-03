import { Account, RpcProvider, ec, uint256, constants } from 'starknet'

export const ERC20_ADDRESS = '0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7'

export const DEVNET_URL = 'http://127.0.0.1:5050'
const DEVNET_NAME = 'devnet'
// This function loads options from the environment.
// It returns options for Devnet as default when nothing is configured in the environment.
export const makeFunderOptsFromEnv = () => {
  const network = process.env.NETWORK || DEVNET_NAME
  const gateway = process.env.NODE_URL || DEVNET_URL
  const accountAddr = process.env.ACCOUNT?.toLowerCase()
  const keyPair = ec.starkCurve.utils.randomPrivateKey()

  return { network, gateway, accountAddr, keyPair }
}

interface FundAccounts {
  account: string
  amount: number
}

interface FunderOptions {
  network?: string
  gateway?: string
  accountAddr?: string
  keyPair: Uint8Array
}

// Define the Strategy to use depending on the network.
export class Funder {
  private opts: FunderOptions
  private strategy: IFundingStrategy

  constructor(opts: FunderOptions) {
    this.opts = opts
    if (this.opts.network === DEVNET_NAME) {
      this.strategy = new DevnetFundingStrategy()
      return
    }
    this.strategy = new AllowanceFundingStrategy()
  }

  // This function adds some funds to pre-deployed account that we are using in our test.
  public async fund(accounts: FundAccounts[]) {
    await this.strategy.fund(accounts, this.opts)
  }
}

interface IFundingStrategy {
  fund(accounts: FundAccounts[], opts: FunderOptions): Promise<void>
}

// Fund the Account on Devnet
class DevnetFundingStrategy implements IFundingStrategy {
  public async fund(accounts: FundAccounts[], opts: FunderOptions) {
    accounts.forEach(async (account) => {
      const body = {
        address: account.account,
        amount: account.amount,
        lite: true,
      }
      await fetch(`${opts.gateway}/mint`, {
        method: 'post',
        body: JSON.stringify(body),
        headers: { 'Content-Type': 'application/json' },
      })
    })
  }
}

// Fund the Account on Testnet
class AllowanceFundingStrategy implements IFundingStrategy {
  public async fund(accounts: FundAccounts[], opts: Required<FunderOptions>) {
    const provider = new RpcProvider({
      nodeUrl: constants.NetworkName.SN_SEPOLIA,
    })

    const operator = new Account(provider, opts.accountAddr, opts.keyPair)

    for (const account of accounts) {
      const data = [
        account.account,
        uint256.bnToUint256(account.amount).low.toString(),
        uint256.bnToUint256(account.amount).high.toString(),
      ]
      const nonce = await operator.getNonce()
      const hash = await operator.execute(
        {
          contractAddress: ERC20_ADDRESS,
          entrypoint: 'transfer',
          calldata: data,
        },
        undefined,
        { nonce },
      )
      await provider.waitForTransaction(hash.transaction_hash)
    }
  }
}
