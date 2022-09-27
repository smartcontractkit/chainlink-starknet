# Starknet On-chain monitor

## Local development

To run the process:

```bash
STARKNET_RPC_ENDPOINT="http://localhost:5050" \
STARKNET_NETWORK_NAME="devnet" \
STARKNET_NETWORK_ID="1" \
STARKNET_CHAIN_ID="1" \
STARKNET_READ_TIMEOUT="5s" \
STARKNET_POLL_INTERVAL="10s" \
STARKNET_LINK_TOKEN_ADDRESS="0x068e99bad58ed47878b780e537faff632a8d616a25477bdebf2b30d1ff845a53" \
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
FEEDS_URL="http://localhost:4000/stom-starknet-devnet.json" \
NODES_URL="http://localhost:4000/nodes-starknet-devnet.json" \
go run ./cmd/monitoring/main.go
```

You need a feed configuration file served at `http://localhost:4000/stom-starknet-devnet.json` that looks like this:

```json
[
  {
    "name": "LINK / USD",
    "path": "link-usd-testing",
    "symbol": "$",
    "heartbeat": 0,
    "contract_type": "numerical_median_feed",
    "status": "testing",
    "multiply": "100000000",
    "contract_address": "0x050eb3095c9324d9bcd4b6b44b0b2c40cefa46407bd6ada154467c66581838b0",
    "proxy_address": "0x050eb3095c9324d9bcd4b6b44b0b2c40cefa46407bd6ada154467c66581838b0"
  }
]
```

... and a nodes configuration file served at `http://localhost:4000/nodes-starknet-devnet.json` that looks like this:

```json
[
  {
    "id": "ocr2-internal-0",
    "website": "",
    "name": "OCR2 internal 0",
    "status": "active",
    "nodeAddress": [
      "0x01e0c41664aedd2072676290d0b1dea3534c32b92036817b41453b768c31aa90"
    ],
    "oracleAddress": "0x0000000000000000000000000000000000000000",
    "csaKeys": [
      {
        "nodeName": "OCR2 internal 0",
        "nodeAddress": "0x01e0c41664aedd2072676290d0b1dea3534c32b92036817b41453b768c31aa90",
        "publicKey": "7dc4dece3794bd64bb9f6f7c2ebf7c303aa018cf306dc03b134349a8f1fcde97"
      }
    ]
  }
]
```

You also need to run the dependent programs. There is [docker-compose file in chainlink-relayer](https://github.com/smartcontractkit/chainlink-relay/blob/main/ops/monitoring/docker-compose.yml). In that folder:

```bash
docker-compose up
```
