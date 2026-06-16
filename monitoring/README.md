# STOM

Starknet on-chain monitoring (STOM) polls OCR2 aggregator and token contracts via
JSON-RPC and exports Prometheus metrics.

## Starknet v0.14.3 / RPC requirements

Starknet v0.14.3 (Sepolia: June 2026, Mainnet: June 2026) deprecates RPC 0.8 and
renames the `"pending"` block tag to `"pre_confirmed"`.

STOM requires an **RPC 0.9 or 0.10.x** endpoint (e.g. `.../rpc/v0_10`). RPC 0.8
URLs will fail after network activation. Self-hosted nodes should run Pathfinder
**≥ v0.22.4**.

### Block tags used by STOM

STOM is read-only. Contract view calls (`starknet_call` for `latest_round_data`,
`link_available_for_payment`, ERC20 `balance_of`, etc.) use the **`latest`** block
tag — the most recent block finalized by L2 consensus. This keeps metrics stable
and aligned with finalized on-chain state.

The relayer TXM (transaction manager) uses **`pre_confirmed`** separately for
nonce lookup and fee estimation, where in-flight transaction state is required.
STOM does not use those code paths.

Event queries use a block number when available, and only fall back to
`pre_confirmed` when fetching events for a block that is still ahead of the
chain tip.

## Useful links

- Starknet on-chain monitor [generated docs](https://pkg.go.dev/github.com/smartcontractkit/chainlink-starknet/monitoring/pkg/monitoring).
- On-chain monitoring (OM) framework architecture docs in [blueprints](https://github.com/smartcontractkit/chainlink-blueprints/blob/master/monitoring/README.md).
- OM framework [generated docs](https://pkg.go.dev/github.com/smartcontractkit/chainlink-common/pkg/monitoring).

## Local development

Note: Previously, this monitor also wrote to Kafka, but the dependency on Kafka has been removed in order to simplify deployment. The kafka topics were unused anyway.

- Start an http server that mimics weiwatchers locally. It needs to export a json configuration file for feeds:

```json
[
  {
    "name": "LINK / USD",
    "path": "link-usd",
    "symbol": "$",
    "heartbeat": 0,
    "contract_type": "numerical_median_feed",
    "status": "testing",
    "contract_address": "<CONTRACT_ADDRESS>",
    "multiply": "100000000",
    "proxy_address": "<PROXY_ADDRESS>"
  }
]
```

It also needs to export a json configuration for for node operators:

```json
[
  {
    "id": "noop",
    "nodeAddress": [<NODE_OPERATOR_ADDRESS>]
  }
]
```

One option is to create a folder `/tmp/configs` and add two files `feeds.json` and `nodes.json` with the configs from above, then:

```bash
python3 -m http.server 4000
```

- Start STOM locally. You will need an RPC 0.9 or 0.10.x endpoint (e.g. `.../rpc/v0_10`) and the address of the LINK token. Make sure you `cd ./monitoring`.

```bash
STARKNET_RPC_ENDPOINT="<RPC_ENDPOINT>" \
STARKNET_NETWORK_NAME="devnet" \
STARKNET_NETWORK_ID="1" \
STARKNET_CHAIN_ID="1" \
STARKNET_READ_TIMEOUT="5s" \
STARKNET_POLL_INTERVAL="10s" #test \
STARKNET_LINK_TOKEN_ADDRESS="<LINK_TOKEN_ADDRESS>" \
HTTP_ADDRESS="localhost:3000" \
FEEDS_URL="http://localhost:4000/feeds.json" \
NODES_URL="http://localhost:4000/nodes.json" \
go run ./cmd/monitoring/main.go
```

- Check the output for the Prometheus scraper

```bash
curl http://localhost:3000/metrics
```
