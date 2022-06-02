import "@shardlabs/starknet-hardhat-plugin";

import { HardhatUserConfig} from "hardhat/types";

const config: HardhatUserConfig = {
  solidity: '0.6.12',
  starknet: {
    venv: "active",
    wallets: {
      OpenZeppelin: {
        accountName: "OpenZeppelin",
        modulePath: "starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount",
        accountPath: "~/.starknet_accounts"
      }
    }
  },
  // paths: {
  //   // starknetSources: __dirname + "/src/starkware/starknet/apps/starkgate/cairo",
  //   cairoPaths: ["../"]
  // },
  networks: {
    devnet: {
      url: "http://127.0.0.1:5050"
    },
  }
};

export default config;