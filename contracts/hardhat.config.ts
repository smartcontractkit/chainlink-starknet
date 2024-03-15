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
  // NOTE: hardhat comes with a built-in special network called 'harhdat'. This network is automatically created and
  // used if no networks are defined in our config: https://hardhat.org/hardhat-runner/docs/config#hardhat-network. We
  // do NOT want to use this network. Our testing scripts already spawn a hardhat node in a container, and we want to
  // use this for the l1 <> l2 messaging tests rather than the automatically generated network. With that in mind, we
  // need to modify this config to point to the hardhat container by adding the 'defaultNetwork' and 'networks' properties.
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
