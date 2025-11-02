[//]: # (Documentation generated from docs.toml - DO NOT EDIT.)
This document describes the TOML format for configuration.
## Example

```toml
ChainID = '<id>'

[[Nodes]]
Name = 'primary'
URL = '<http url>'
APIKey = '<key>'
```

## Global
```toml
ChainID = 'Ibiza-808' # Example
FeederURL = 'http://feeder.url' # Example
Enabled = true # Default
OCR2CachePollPeriod = '5s' # Default
OCR2CacheTTL = '1m' # Default
RequestTimeout = '10s' # Default
TxTimeout = '10s' # Default
ConfirmationPoll = '5s' # Default
FeeEstimationMaxAttempts = 5 # Default
```


### ChainID
```toml
ChainID = 'Ibiza-808' # Example
```
ChainID is the Starknet chain ID.

### FeederURL
```toml
FeederURL = 'http://feeder.url' # Example
```
FeederURL is required to get tx metadata (that the RPC can't)

### Enabled
```toml
Enabled = true # Default
```
Enabled enables this chain.

### OCR2CachePollPeriod
```toml
OCR2CachePollPeriod = '5s' # Default
```
OCR2CachePollPeriod is the rate to poll for the OCR2 state cache.

### OCR2CacheTTL
```toml
OCR2CacheTTL = '1m' # Default
```
OCR2CacheTTL is the stale OCR2 cache deadline.

### RequestTimeout
```toml
RequestTimeout = '10s' # Default
```
RequestTimeout is the RPC client timeout.

### TxTimeout
```toml
TxTimeout = '10s' # Default
```
TxTimeout is the timeout for sending txes to an RPC endpoint.

### ConfirmationPoll
```toml
ConfirmationPoll = '5s' # Default
```
ConfirmationPoll is how often to confirmer checks for tx inclusion on chain.

### FeeEstimationMaxAttempts
```toml
FeeEstimationMaxAttempts = 5 # Default
```
FeeEstimationMaxAttempts is the maximum number of retry attempts for fee estimation.

## Nodes
```toml
[[Nodes]]
Name = 'primary' # Example
URL = 'http://stark.node' # Example
APIKey = 'key' # Example
```


### Name
```toml
Name = 'primary' # Example
```
Name is a unique (per-chain) identifier for this node.

### URL
```toml
URL = 'http://stark.node' # Example
```
URL is the base HTTP(S) endpoint for this node.

### APIKey
```toml
APIKey = 'key' # Example
```
APIKey Header is optional and only required for Nethermind RPCs

