## Integration tests

### Run tests

1. `ginkgo -r --focus @ocr integration-tests/smoke` (from core of repo)

### Structure

[Commons](../../integration-tests/common/common.go) - Common EVM based methods to generate chains, nodes, key bundles

[Test Commons](../../integration-tests/common/test_common.go) - Test methods to deploy env, configure clients, fetch client details

[Starknet Commons](../../ops/devnet/devnet.go) - Methods related to starknet and L2 actions such as minting, L1<>L2 sync

[Gauntlet wrapper](../../relayer/pkg/starknet/gauntlet_starknet.go) - Wrapper for Starknet gauntlet

[OCRv2 tests](../../integration-tests/smoke/ocr2_test.go) - Example smoke test to set up environment, configure it and run the smoke test

### Writing tests

See examples [here](../../integration-tests/smoke/ocr2_test.go)

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
