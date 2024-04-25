package monitoring

import (
	"context"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/felt"
	starknetutils "github.com/NethermindEth/starknet.go/utils"
	commonMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/erc20"
)

type ContractAddress struct {
	Address *felt.Felt
	Name    string
}

type ContractAddressWithBalance struct {
	ContractAddress
	Balance *big.Int
}

type BalanceEnvelope struct {
	Contracts []ContractAddressWithBalance
	Decimals  *big.Int
}

func NewContractAddress(address string, name string) (ContractAddress, error) {
	ans := ContractAddress{}
	addr, err := starknetutils.HexToFelt(address)
	if err != nil {
		return ans, fmt.Errorf("error parsing contract address: %w", err)
	}
	ans.Address = addr
	ans.Name = name
	return ans, nil
}

type nodeBalancesSourceFactory struct {
	erc20Reader erc20.ERC20Reader
}

func NewNodeBalancesSourceFactory(erc20Reader erc20.ERC20Reader) *nodeBalancesSourceFactory {
	return &nodeBalancesSourceFactory{
		erc20Reader: erc20Reader,
	}
}

func (f *nodeBalancesSourceFactory) NewSource(
	_ commonMonitoring.ChainConfig,
	rddNodes []commonMonitoring.NodeConfig,
) (commonMonitoring.Source, error) {
	var addrs []ContractAddress

	for _, n := range rddNodes {
		addr, err := NewContractAddress(string(n.GetAccount()), n.GetName())
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return &contractBalancesSource{erc20Reader: f.erc20Reader, contracts: addrs}, nil
}

func (f *nodeBalancesSourceFactory) GetType() string {
	return "nodeBalances"
}

// contract balances sources can be potentially reused for other contracts (not just the node account contracts)
type contractBalancesSource struct {
	erc20Reader erc20.ERC20Reader
	contracts   []ContractAddress
}

func (s *contractBalancesSource) Fetch(ctx context.Context) (interface{}, error) {
	var cAns []ContractAddressWithBalance
	for _, c := range s.contracts {
		balance, err := s.erc20Reader.BalanceOf(ctx, c.Address)
		if err != nil {
			return nil, fmt.Errorf("could not fetch address balance %w", err)
		}
		cAns = append(cAns, ContractAddressWithBalance{c, balance})
	}
	dAns, err := s.erc20Reader.Decimals(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch decimals %w", err)
	}

	return BalanceEnvelope{Contracts: cAns, Decimals: dAns}, nil
}
