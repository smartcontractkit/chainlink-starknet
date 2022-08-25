import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'
import '@nomiclabs/hardhat-ethers'
import '@nomiclabs/hardhat-waffle'

const COMPILER_SETTINGS = {
  optimizer: {
    enabled: true,
    runs: 1000000,
  },
  metadata: {
    bytecodeHash: 'none',
  },
}

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
const config: HardhatUserConfig = {
  solidity: {
    compilers: [
      {
        version: '0.6.12',
        settings: COMPILER_SETTINGS,
      },
      {
        version: '0.8.15',
        settings: COMPILER_SETTINGS,
      },
    ],
  },
  starknet: {
    // dockerizedVersion: "0.8.1", // alternatively choose one of the two venv options below
    // uses (my-venv) defined by `python -m venv path/to/my-venv`
    // venv: "../.venv",

    // uses the currently active Python environment (hopefully with available Starknet commands!)
    venv: 'active',
    // network: "alpha",
    network: 'devnet',
    wallets: {
      OpenZeppelin: {
        accountName: 'OpenZeppelin',
        modulePath: 'starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount',
        accountPath: '~/.starknet_accounts',
      },
    },
  },
  networks: {
    devnet: {
      url: 'http://127.0.0.1:5050',
    },
    integratedDevnet: {
      url: 'http://127.0.0.1:5050',
      venv: 'active',
      args: ['--lite-mode'],
      // dockerizedVersion: "0.2.0"
    },
  },
  mocha: {
    timeout: 10000000,
  },
  paths: {
    sources: './src',
    cairoPaths: ['./src', './vendor/starkware-libs/starkgate-contracts/src'],
  },
}

export default config
