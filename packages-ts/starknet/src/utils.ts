import { constants, encode, number, Account, SequencerProvider, ec } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { expect } from 'chai'
import { artifacts, network } from 'hardhat'

export const ERC20_ADDRESS_DEVNET =
  '0x62230ea046a9a5fbc261ac77d03c8d41e5d442db2284587570ab46455fd2488'
export const ERC20_ADDRESS_TESTNET =
  '0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7'

// This function adds the build info to the test network so that the network knows
// how to handle custom errors.  It is automatically done when testing
// against the default hardhat network.
export const addCompilationToNetwork = async (fullyQualifiedName: string) => {
  if (network.name !== 'hardhat') {
    // This is so that the network can know about custom errors.
    // Running against the provided hardhat node does this automatically.

    const buildInfo = await artifacts.getBuildInfo(fullyQualifiedName)
    if (!buildInfo) {
      throw Error('Cannot find build info')
    }
    const { solcVersion, input, output } = buildInfo
    console.log('Sending compilation result for StarkNetValidator test')
    await network.provider.request({
      method: 'hardhat_addCompilationResult',
      params: [solcVersion, input, output],
    })
    console.log('Successfully sent compilation result for StarkNetValidator test')
  }
}

export const expectInvokeError = async (invoke: Promise<any>, expected?: string) => {
  try {
    await invoke
  } catch (err: any) {
    expectInvokeErrorMsg(err?.message, expected)
    return // force
  }
  expect.fail("Unexpected! Invoke didn't error!?")
}

export const expectInvokeErrorMsg = (actual: string, expected?: string) => {
  // Match transaction error
  expect(actual).to.deep.contain('Transaction rejected. Error message:')
  // Match specific error
  if (expected) expectSpecificMsg(actual, expected)
}

export const expectCallError = async (call: Promise<any>, expected?: string) => {
  try {
    await call
  } catch (err: any) {
    expectCallErrorMsg(err?.message, expected)
    return // force
  }
  expect.fail("Unexpected! Call didn't error!?")
}

export const expectCallErrorMsg = (actual: string, expected?: string) => {
  // Match call error
  expect(actual).to.deep.contain('Could not perform call')
  // Match specific error
  if (expected) expectSpecificMsg(actual, expected)
}

export const expectSpecificMsg = (actual: string, expected: string) => {
  // Match specific error
  const matches = actual.match(/Error message: (.+?)\n/g)
  // Joint matches should include the expected, or fail
  if (matches && matches.length > 0) {
    expect(matches.join()).to.include(expected)
  } else expect.fail(`\nActual: ${actual}\n\nExpected: ${expected}`)
}

// Required to convert negative values into [0, PRIME) range
export const toFelt = (int: number | BigNumberish): BigNumberish => {
  const prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME))
  return number.toBN(int).umod(prime)
}

// NOTICE: Leading zeros are trimmed for an encoded felt (number).
//   To decode, the raw felt needs to be start padded up to max felt size (252 bits or < 32 bytes).
export const hexPadStart = (data: number | bigint, len: number) => {
  return `0x${data.toString(16).padStart(len, '0')}`
}

interface FunderOpts {
  makeFunderOptsFromEnv: () => AccountFunderOptions
  Funder: (opts: AccountFunderOptions) => AccountFunder
}

// This function loads options from the environment.
// It returns options for Devnet as default when nothing is configured in the environment.
const makeFunderOptsFromEnv = (): AccountFunderOptions => {
  const network = process.env.NETWORK || 'devnet'
  const gateway_url = process.env.NODE_URL || 'http://127.0.0.1:5050'

  return { network, gateway_url }
}

const Funder = (opts: AccountFunderOptions): AccountFunder => {
  return new AccountFunder(opts)
}

export const account: FunderOpts = {
  makeFunderOptsFromEnv: makeFunderOptsFromEnv,
  Funder: Funder,
}

interface FundAccounts {
  account: string
  amount: number
}

interface AccountFunderOptions {
  network?: string
  gateway_url?: string
}

// Define the Strategy to use depending on the network.
class AccountFunder {
  private opts: AccountFunderOptions
  private strategy: IFundingStrategy

  constructor(opts: AccountFunderOptions) {
    this.opts = opts
    if (this.opts.network === 'devnet') {
      this.strategy = new DevnetFundingStrategy()
      return
    }
    this.strategy = new AllowanceFundingStrategy()
  }

  //This function add some funds to predeploy account that we are using in our test.
  public async fund(accounts: FundAccounts[]) {
    this.strategy.fund(accounts, this.opts)
  }
}

interface IFundingStrategy {
  fund(accounts: FundAccounts[], opts: AccountFunderOptions): void
}

// Fund the Account on Devnet
class DevnetFundingStrategy implements IFundingStrategy {
  public async fund(accounts: FundAccounts[], opts: AccountFunderOptions) {
    accounts.forEach(async (account) => {
      const body = {
        address: account.account,
        amount: account.amount,
        lite: true,
      }

      try {
        await fetch(`${opts.gateway_url}/mint`, {
          method: 'post',
          body: JSON.stringify(body),
          headers: { 'Content-Type': 'application/json' },
        })
      } catch (error: any) {}
    })
  }
}

// Fund the Account on Testnet
class AllowanceFundingStrategy implements IFundingStrategy {
  public async fund(accounts: FundAccounts[], opts: AccountFunderOptions) {
    const provider = new SequencerProvider({
      baseUrl: 'https://alpha4.starknet.io',
    })

    const keyPair = ec.getKeyPair(process.env.ACCOUNT_PRIVATE_KEY)
    const accountFunder = new Account(provider, process.env.ACCOUNT.toLowerCase(), keyPair)

    accounts.forEach(async (account) => {
      const hash = await accountFunder.execute(
        {
          contractAddress: ERC20_ADDRESS_TESTNET,
          entrypoint: 'transfer',
          calldata: [account.account, account.amount],
        },
        undefined,
        { maxFee: '32703804275172' },
      )
      await provider.waitForTransaction(hash.transaction_hash)
    })
  }
}
