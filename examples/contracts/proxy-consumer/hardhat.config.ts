import { HardhatUserConfig } from 'hardhat/types'
import '@shardlabs/starknet-hardhat-plugin'

const config: HardhatUserConfig = {
  solidity: '0.8.14',
  starknet: {
    venv: 'active',
    network: 'alphaGoerli',
    wallets: {
      OpenZeppelin: {
        accountName: 'OpenZeppelin',
        modulePath: 'starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount',
        accountPath: '~/.starknet_accounts',
      },
    },
  },
  paths: {
    cairoPaths: ['../../../contracts/src'],
  },
}

export default config
