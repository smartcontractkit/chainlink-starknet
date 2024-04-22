package erc20

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

//go:generate mockery --name ERC20Reader --output ./mocks/
type ERC20Reader interface {
	BalanceOf(context.Context, *felt.Felt) (*big.Int, error)
	Decimals(context.Context) (*big.Int, error)

	BaseReader() starknet.Reader
}

var _ ERC20Reader = (*Client)(nil)

type Client struct {
	r            starknet.Reader
	lggr         logger.Logger
	tokenAddress *felt.Felt
	decimals     *big.Int
}

func NewClient(reader starknet.Reader, lggr logger.Logger, tokenAddress *felt.Felt) (*Client, error) {
	return &Client{
		r:            reader,
		lggr:         lggr,
		tokenAddress: tokenAddress,
		// lazy initialize decimals
		decimals: nil,
	}, nil
}

func (c *Client) BalanceOf(ctx context.Context, accountAddress *felt.Felt) (*big.Int, error) {
	ops := starknet.CallOps{
		ContractAddress: c.tokenAddress,
		Selector:        starknetutils.GetSelectorFromNameFelt("balance_of"),
		Calldata:        []*felt.Felt{accountAddress},
	}

	balanceRes, err := c.r.CallContract(ctx, ops)
	if err != nil {
		return nil, fmt.Errorf("couldn't call balance_of on erc20: %w", err)
	}

	return starknetutils.FeltToBigInt(balanceRes[0]), nil

}

func (c *Client) Decimals(ctx context.Context) (*big.Int, error) {
	if c.decimals == nil {

		ops := starknet.CallOps{
			ContractAddress: c.tokenAddress,
			Selector:        starknetutils.GetSelectorFromNameFelt("decimals"),
		}

		decimalsRes, err := c.r.CallContract(ctx, ops)

		if err != nil {
			return nil, fmt.Errorf("couldn't call decimals on erc20: %w", err)
		}

		c.decimals = starknetutils.FeltToBigInt(decimalsRes[0])
	}

	return c.decimals, nil
}

func (c *Client) BaseReader() starknet.Reader {
	return c.r
}
