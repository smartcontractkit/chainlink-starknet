## Integration tests

### Run tests

### Prerequisites

1. `yarn install`
2. `yarn build`

#### Smoke

`cd integration-tests/smoke/ && go test --timeout=2h -v` (from core of repo)

#### Soak

Soak tests will run a modified version of the smoke test via a remote runner for the set period. The difference is that
there is no panic when an
error appears, but instead log it.

##### Run

`make test-integration-soak`

##### Env vars

`TTL=72h` - duration of soak

`NODE_COUNT` - number of OCR nodes

`CHAINLINK_IMAGE` - Chainlink docker image repo

`CHAINLINK_VERSION` - Chainlink docker image version

`L2_RPC_URL` - This will override the L2 url, used for testnet (optional)

`PRIVATE_KEY` - Private key for Testnet (optional)

`ACCOUNT` - Account address on Testnet (optional)

### Structure

[Commons](../../integration-tests/common/common.go) - Common Chainlink methods to generate chains, nodes, key bundles

[Test Commons](../../integration-tests/common/test_common.go) - Test methods to deploy env, configure clients, fetch
client details

[Starknet Commons](../../ops/devnet/devnet.go) - Methods related to starknet and L2 actions such as minting, L1<>L2 sync

[Gauntlet wrapper](../../relayer/pkg/starknet/gauntlet_starknet.go) - Wrapper for Starknet gauntlet

[OCRv2 tests](../../integration-tests/smoke/ocr2_test.go) - Example smoke test to set up environment, configure it and
run the smoke test

### Writing tests

See smoke examples [here](../../integration-tests/smoke/ocr2_test.go)

See soak examples [here](../../integration-tests/soak/tests/ocr_test.go)
and [here](../../integration-tests/soak/soak_runner_test.go)

1. Instantiate Gauntlet
2. Deploy Cluster
3. Set Gauntlet network
4. Deploy accounts on L2 for the nodes
5. Fund the accounts
6. Deploy L2 LINK token via Gauntlet
7. Deploy L2 Access controller contract via Gauntlet
8. Deploy L2 OCR2 contract via Gauntlet
9. Set OCR2 billing via Gauntlet
10. Set OCR2 config details via Gauntlet
11. Set up boostrap and oracle nodes

### Metrics and logs (K8)

1. Navigate to Grafana
2. Search for `chainlink-testing-insights` dashboard
3. Select the starknet namespace

Here you will find pod logs for all the chainlink nodes as well as Devnet / Geth

# Testing wiki

## Testnet

- Chain name - `Starknet`
- Chain ID - `SN_GOERLI`
  - Testnet 1 - `[https://alpha4.starknet.io](https://alpha4.starknet.io)`
  - Testnet 2 - [`https://alpha4-2.starknet.io`](https://alpha4-2.starknet.io/)

## Mainnet

- Chain name - `Starknet`
- Chain ID - `SN_MAIN`
  - `[https://alpha-mainnet.starknet.io](https://alpha-mainnet.starknet.io)`

# Node config

```bash
[[Starknet]]
Enabled = true
ChainID = '<id>'
[[Starknet.Nodes]]
Name = 'primary'
URL = '<rpc>'

[OCR2]
Enabled = true

[P2P]
[P2P.V2]
Enabled = true
DeltaDial = '5s'
DeltaReconcile = '5s'
ListenAddresses = ['0.0.0.0:6690']
```

# Gauntlet steps

## Environment file

```bash
NODE_URL=<rpc_url>
ACCOUNT=<account>
PRIVATE_KEY=<private_key>
```

1. Deploy link

```bash
yarn gauntlet token:deploy --link
```

2. Deploy access controller

```bash
yarn gauntlet access_controller:deploy
```

3. Deploy OCR2

```bash
yarn gauntlet ocr2:deploy --minSubmissionValue=<value> --maxSubmissionValue=<value> --decimals=<value> --name=<value> --link=<link_addr>
```

4. Deploy proxy

```bash
yarn gauntlet proxy:deploy <ocr_address>
```

5. Add access to proxy

```bash
yarn gauntlet ocr2:add_access --address=<ocr_address> <proxy_address>
```

6. Mint LINK

```bash
yarn gauntlet token:mint --recipient<ocr_addr> --amount=<value> <link_addr>
```

7. Set billing

```bash
yarn gauntlet ocr2:set_billing --observationPaymentGjuels=<value> --transmissionPaymentGjuels=<value> <ocr_addr>
```

8. Set config

   1. Example config testnet

   ```bash
   {
       "f": 1,
       "signers": [
           "ocr2on_starknet_0371028377bfd793b7e2965757e348309e7242802d20253da6ab81c8eb4b4051",
           "ocr2on_starknet_073cadfc4474e8c6c79f66fa609da1dbcd5be4299ff9b1f71646206d1faca1fc",
           "ocr2on_starknet_0386d1a9d93792c426739f73afa1d0b19782fbf30ae27ce33c9fbd4da659cd80",
           "ocr2on_starknet_005360052758819ba2af790469a28353b7ff6f8b84176064ab572f6cc20e5fb4"
       ],
       "transmitters": [
           "0x0...",
           "0x0...",
           "0x0...",
           "0x0..."
       ],
       "onchainConfig": "",
       "offchainConfig": {
           "deltaProgressNanoseconds": 8000000000,
           "deltaResendNanoseconds": 30000000000,
           "deltaRoundNanoseconds": 3000000000,
           "deltaGraceNanoseconds": 1000000000,
           "deltaStageNanoseconds": 20000000000,
           "rMax": 5,
           "s": [
               1,
               1,
               1,
               1
           ],
           "offchainPublicKeys": [
               "ocr2off_starknet_0...",
               "ocr2off_starknet_0...",
               "ocr2off_starknet_0...",
               "ocr2off_starknet_0..."
           ],
           "peerIds": [
               "12D3..",
               "12D3..",
               "12D3..",
               "12D3.."
           ],
           "reportingPluginConfig": {
               "alphaReportInfinite": false,
               "alphaReportPpb": 0,
               "alphaAcceptInfinite": false,
               "alphaAcceptPpb": 0,
               "deltaCNanoseconds": 1000000000
           },
           "maxDurationQueryNanoseconds": 2000000000,
           "maxDurationObservationNanoseconds": 1000000000,
           "maxDurationReportNanoseconds": 2000000000,
           "maxDurationShouldAcceptFinalizedReportNanoseconds": 2000000000,
           "maxDurationShouldTransmitAcceptedReportNanoseconds": 2000000000,
           "configPublicKeys": [
               "ocr2cfg_starknet_...",
               "ocr2cfg_starknet_...",
               "ocr2cfg_starknet_...",
               "ocr2cfg_starknet_..."
           ]
       },
       "offchainConfigVersion": 2,
       "secret": "some secret you want"
   }
   ```

```bash
yarn gauntlet ocr2:set_config --input=<cfg> <ocr_addr>
```
