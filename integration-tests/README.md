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

#### Run smoke tests

`cd integration-tests && go test --timeout=2h -v -count=1 -json ./smoke`


### On demand soak test

Navigate to the [workflow](https://github.com/smartcontractkit/chainlink-starknet/actions/workflows/integration-tests-soak.yml). The workflow takes in 3 parameters:

- Base64 string of the .toml configuration
- Core image tag which defaults to develop
- Test runner tag, only tag needs to be supplied

Create an `overrides.toml` file in `integration-tests/testconfig` and run `cat overrides.toml | base64`. `inside_k8` needs to be set to true in the .toml in order to run the tests in kubernetes.

#### Local

If you want to kick off the test from local:

- `export TEST_SUITE: soak`
- `export DETACH_RUNNER: true`
- `export ENV_JOB_IMAGE: <internal_repo>/chainlink-solana-tests:<tag>`
- Base64 the .toml config
- Run `export BASE64_CONFIG_OVERRIDE="<config>"`
- `cd integration-tests/soak && go test -timeout 24h -count=1 -run TestOCRBasicSoak/embedded -test.timeout 30m;`


