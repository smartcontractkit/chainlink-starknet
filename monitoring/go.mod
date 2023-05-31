module github.com/smartcontractkit/chainlink-starknet/monitoring

go 1.20

require (
	github.com/prometheus/client_golang v1.14.0
	github.com/smartcontractkit/caigo v0.0.0-20230526231506-786d4587099a
	github.com/smartcontractkit/chainlink-relay v0.1.7-0.20230531014621-9c303da4c086
	github.com/smartcontractkit/chainlink-starknet/relayer v0.0.0-20230508053614-9f2fd5fd4ff1
	github.com/smartcontractkit/libocr v0.0.0-20230525150148-a75f6e244bb3
	github.com/stretchr/testify v1.8.2
	go.uber.org/multierr v1.9.0
)

require (
	github.com/NethermindEth/juno v0.0.0-20220630151419-cbd368b222ac // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/confluentinc/confluent-kafka-go v1.9.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/ethereum/go-ethereum v1.11.5 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-plugin v1.4.9 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.12.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-testing-interface v0.0.0-20171004221916-a61a99592b77 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/riferrei/srclient v0.5.4 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.1.1 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.5.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20230425010034-47ecfdc1ba53 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230223222841-637eb2293923 // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	// Fix go mod tidy issue for ambiguous imports from go-ethereum
	// See https://github.com/ugorji/go/issues/279
	github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1

	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	github.com/smartcontractkit/chainlink-starknet/relayer => ../relayer
)
