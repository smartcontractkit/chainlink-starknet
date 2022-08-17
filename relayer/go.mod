module github.com/smartcontractkit/chainlink-starknet/relayer

go 1.18

require (
	github.com/NethermindEth/juno v0.0.0-20220630151419-cbd368b222ac
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1
	github.com/dontpanicdao/caigo v0.3.1-0.20220812122711-b855f2b57bb5
	github.com/pkg/errors v0.9.1
	github.com/smartcontractkit/chainlink-relay v0.1.5-0.20220808181113-70f8468a87ee
	github.com/smartcontractkit/chainlink-starknet/ops v0.0.0-20220818192054-2a761cdd6f6a
	github.com/smartcontractkit/libocr v0.0.0-20220701150323-d815c8d0eab8
	github.com/stretchr/testify v1.8.0
	github.com/test-go/testify v1.1.4
	golang.org/x/exp v0.0.0-20220608143224-64259d1afd70
	gopkg.in/guregu/null.v4 v4.0.0
)

require (
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/go-ethereum v1.10.21 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/zerolog v1.27.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/smartcontractkit/chainlink-testing-framework v1.5.8 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/smartcontractkit/chainlink-starknet/ops => ../../chainlink-starknet/ops
