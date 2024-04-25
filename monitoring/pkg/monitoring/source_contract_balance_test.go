package monitoring

import (
	"context"
	"math/big"
	"testing"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	erc20Mocks "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/erc20/mocks"
)

func TestContractBalancesSource(t *testing.T) {
	// This test makes sure that the mapping between the response from the ocr2.Client
	// method calls and the output of the Proxy source is correct.

	chainConfig := generateChainConfig()
	nodeConfig := generateNodeConfig()

	// ocr2Reader := ocr2Mocks.NewOCR2Reader(t)
	// ocr2Reader.On(
	// 	"LatestRoundData",
	// 	mock.Anything, // ctx
	// 	proxyContractAddressFelt,
	// ).Return(ocr2ClientLatestRoundDataResponseForProxy, nil).Once()

	erc20Reader := erc20Mocks.NewERC20Reader(t)

	for _, x := range nodeConfig {
		nodeAddressFelt, err := starknetutils.HexToFelt(string(x.GetAccount()))
		require.NoError(t, err)

		erc20Reader.On(
			"BalanceOf",
			mock.Anything,   // ctx
			nodeAddressFelt, // address
		).Return(new(big.Int).SetUint64(777), nil)
	}

	erc20Reader.On(
		"Decimals",
		mock.Anything, // ctx
	).Return(new(big.Int).SetUint64(18), nil)

	factory := NewNodeBalancesSourceFactory(erc20Reader)
	source, err := factory.NewSource(chainConfig, nodeConfig)
	require.NoError(t, err)
	rawBalanceEnvelope, err := source.Fetch(context.Background())
	require.NoError(t, err)
	balanceEnvelope, ok := rawBalanceEnvelope.(BalanceEnvelope)
	require.True(t, ok)

	require.Equal(t, balanceEnvelope.Decimals.Uint64(), uint64(18))

	require.Equal(t, balanceEnvelope.Contracts[0].Balance.Uint64(), uint64(777))
	require.Equal(t, balanceEnvelope.Contracts[1].Balance.Uint64(), uint64(777))

}
