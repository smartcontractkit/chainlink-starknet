## Integration tests - HOWTO

### Prerequisites
1. `cd contracts && scarb --profile release build`
2. `yarn install`
3. `yarn build`

#### TOML preparation
The integration tests are using TOML as the configuration input. The logic and parsing is located under [Test config](./testconfig)

By default, the tests will be running with the default config set in [default.toml](./testconfig/default.toml). This configuration is set to run on devnet with local docker.

Fields in the default toml can be overriden by creating an `overrides.toml`file. Any values specified here take precedence and will be overwritten if they overlap with `default.toml`.

##### Testnet runs
In order to run the tests on Testnet, additional variables need to be specified in the TOML, these would also be pointed out if `network = "testnet"` is set. The additional variables are:

- `l2_rpc_url` - L2 RPC url
- `account` - Account address on L2
- `private_key` - Private key for L2 account

##### Running in k8s

Set `inside_k8 = true` under `[Common]`.

#### Run tests

`cd integration-tests && go test --timeout=2h -v -count=1 -json ./smoke`