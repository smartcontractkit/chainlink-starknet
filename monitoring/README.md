# STOM

## Useful links

- Starknet on-chain monitor [generated docs](https://pkg.go.dev/github.com/smartcontractkit/chainlink-starknet/monitoring/pkg/monitoring).
- On-chain monitoring (OM) framework architecture docs in [blueprints](https://github.com/smartcontractkit/chainlink-blueprints/blob/master/monitoring/README.md).
- OM framework [generated docs](https://pkg.go.dev/github.com/smartcontractkit/chainlink-relay/pkg/monitoring).

## Local development

- Start the third party dependencies using [docker-compose](https://docs.docker.com/compose/).
  Use the docker-compose.yml file from [smartcontractkit/chainlink-relay/ops/monitoring](https://github.com/smartcontractkit/chainlink-relay/tree/main/ops/monitoring).

```sh
docker-compose up
```

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
    "nodeAddress": ["0x26db6cd9e7dfd3f7c825ec9d6d2646e7a959fc574febde9668337e4c55aaac"]
  }
]
```

One option is to create a folder `/tmp/configs` and add two files `feeds.json` and `nodes.json` with the configs from above, then:

```bash
python3 -m http.server 4000
```

- Start STOM locally. You will need and RPC endpoint and the address of the LINK token. Make sure you `cd ./monitoring`.

```bash
STARKNET_RPC_ENDPOINT="<RPC_ENDPOINT>" \
STARKNET_NETWORK_NAME="devnet" \
STARKNET_NETWORK_ID="1" \
STARKNET_CHAIN_ID="1" \
STARKNET_READ_TIMEOUT="5s" \
STARKNET_POLL_INTERVAL="10s" #test \
STARKNET_LINK_TOKEN_ADDRESS="<LINK_TOKEN_ADDRESS>" \
KAFKA_BROKERS="localhost:29092" \
KAFKA_CLIENT_ID="starknet" \
KAFKA_SECURITY_PROTOCOL="PLAINTEXT" \
KAFKA_SASL_MECHANISM="PLAIN" \
KAFKA_SASL_USERNAME="" \
KAFKA_SASL_PASSWORD="" \
KAFKA_CONFIG_SET_SIMPLIFIED_TOPIC="config_set_simplified" \
KAFKA_TRANSMISSION_TOPIC="transmission_topic" \
SCHEMA_REGISTRY_URL="http://localhost:8989" \
SCHEMA_REGISTRY_USERNAME="" \
SCHEMA_REGISTRY_PASSWORD="" \
HTTP_ADDRESS="localhost:3000" \
FEEDS_URL="http://localhost:4000/feeds.json" \
NODES_URL="http://localhost:4000/nodes.json" \
go run ./cmd/monitoring/main.go
```

- Check the output for the Prometheus scraper

```bash
curl http://localhost:3000/metrics
```

- To check the output for Kafka, you need to install [kcat](https://github.com/edenhill/kcat). After you install, run:

```bash
kcat -b localhost:29092 -t config_set_simplified
kcat -b localhost:29092 -t transmission_topic
```
