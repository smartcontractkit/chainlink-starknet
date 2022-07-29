import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'
import '@nomiclabs/hardhat-ethers'
import '@nomiclabs/hardhat-etherscan'
import '@nomiclabs/hardhat-waffle'
import '@typechain/hardhat'
import 'hardhat-abi-exporter'
import 'hardhat-contract-sizer'
import 'solidity-coverage'

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
const config: HardhatUserConfig = {
  // abiExporter: {
  //   path: './abi',
  // },
  // typechain: {
  //   outDir: './typechain',
  //   target: 'ethers-v5',
  // },
  solidity: '0.8.14',
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
  paths: {
    cairoPaths: ['node_modules/@joriksch/oz-cairo/src'],
    artifacts: './artifacts',
    cache: './cache',
    sources: './contracts',
    tests: './test',
  },
  networks: {
    hardhat: {},
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
  // contractSizer: {
  //   alphaSort: true,
  //   runOnCompile: false,
  //   disambiguatePaths: false,
  // },
  mocha: {
    timeout: 100000,
  },
}

export default config
