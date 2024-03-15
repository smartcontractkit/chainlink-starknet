import { HardhatUserConfig } from 'hardhat/types'
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
  // NOTE: hardhat comes with a special built-in network called 'harhdat'. This network is automatically created and
  // used if no networks are defined in our config: https://hardhat.org/hardhat-runner/docs/config#hardhat-network. It
  // is important to note that we DO NOT want to use this network. Our testing scripts already spawn a hardhat node in
  // a container, so we should use this for the l1 <> l2 messaging tests rather than the auto-generated one from hardhat.
  // To achieve this, the 'defaultNetwork' and 'networks' properties have been adjusted such that they reference the
  // containerized hardhat node.
  defaultNetwork: 'localhost',
  networks: {
    localhost: {
      url: 'http://127.0.0.1:8545',
    },
  },
  solidity: {
    compilers: [
      {
        version: '0.8.15',
        settings: COMPILER_SETTINGS,
      },
    ],
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
