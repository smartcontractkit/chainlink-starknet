import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'
import '@nomiclabs/hardhat-ethers'

const config: HardhatUserConfig = {
  solidity: '0.8.14',
  starknet: {
    venv: 'active',
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
  paths: {
    cairoPaths: ['../../contracts/src'],
  },
}

export default config
