import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'
import '@nomiclabs/hardhat-ethers'

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
const config: HardhatUserConfig = {
  solidity: '0.8.14',
  starknet: {
    // dockerizedVersion: "0.8.1", // alternatively choose one of the two venv options below
    // uses (my-venv) defined by `python -m venv path/to/my-venv`
    // venv: "../.venv",

    // uses the currently active Python environment (hopefully with available Starknet commands!)
    venv: 'active',
    // network: "alpha",
    network: 'integrated-devnet',
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
}

export default config
