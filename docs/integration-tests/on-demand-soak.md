## On demand soak test


Soak tests can be triggered in GHA remotely with custom duration on devnet / testnet

1. Navigate to Actions
2. Select Integration Tests - Soak
3. Click run workflow
4. Enter RPC url (Optional this is for testing on testnet)
5. Specify node count (default is 4+1)
6. Specify TTL of the namespace (This is when to destroy the env)
7. Specify duration of the soak (Should be lower than TTL)
8. Enter private key L2 (Optional, only for testnet)
9. Enter account address L2 (Optional, only for testnet)


## Monitoring
Tests will print out a namespace in the "TestOCRSoak" phase in the Run tests step (e.g chainlink-ocr-starknet-472d5)

1. Enter the namespace in grafana chainlink testing insights dashboard and the logs will be visible

The remote runner contains the test run and outputs.