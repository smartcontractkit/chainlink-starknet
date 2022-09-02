package txm

import (
	"math/big"

	"github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/core"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const (
	FEE_MARGIN         uint64 = 115
	AccountCreationErr        = "failed to create account"
)

func New(lggr logger.Logger, keystore keys.Keystore, cfg core.Config, getClient func() (starknet.ReaderWriter, error)) (core.TxManager[types.Transaction], error) {
	return core.New[types.Transaction, keys.Key, *big.Int, *types.Felt](
		lggr,
		keystore,
		cfg,
		&starkChainClient{
			client: utils.NewLazyLoad(getClient),
		},
		&txStatuses{
			txs: map[string]transaction{},
		},
	)
}
