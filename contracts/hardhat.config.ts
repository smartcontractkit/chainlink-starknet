import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'
import '@nomiclabs/hardhat-ethers'
import '@nomicfoundation/hardhat-chai-matchers'
import 'solidity-coverage'
import { prepareHardhatArtifacts } from './test/setup'

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
        version: '0.8.15',
        settings: COMPILER_SETTINGS,
      },
    ],
  },
  starknet: {
    // dockerizedVersion: "0.10.0", // alternatively choose one of the two venv options below
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
    requestTimeout: 1000000,
  },
  networks: {
    devnet: {
      url: 'http://127.0.0.1:5050',
      args: ['--cairo-compiler-manifest', '../vendor/cairo/Cargo.toml'],
    },
    integratedDevnet: {
      url: 'http://127.0.0.1:5050',
      venv: 'active',
      args: ['--lite-mode', '--cairo-compiler-manifest', '../vendor/cairo/Cargo.toml'],
      // dockerizedVersion: "0.2.0"
    },
  },
  mocha: {
    timeout: 10000000,
    rootHooks: {
      beforeAll: prepareHardhatArtifacts,
    },
  },
  paths: {
    sources: './solidity',
    starknetSources: './src',
    starknetArtifacts: './target/release',
    cairoPaths: [],
  },
}

export default config
