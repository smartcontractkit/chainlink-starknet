package monitoring

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	commonMonitoring "github.com/smartcontractkit/chainlink-common/pkg/monitoring"
)

type StarknetNodeConfig struct {
	ID          string   `json:"id,omitempty"`
	NodeAddress []string `json:"nodeAddress,omitempty"`
}

func (s StarknetNodeConfig) GetName() string {
	return s.ID
}

func (s StarknetNodeConfig) GetAccount() types.Account {
	address := ""
	if len(s.NodeAddress) != 0 {
		address = s.NodeAddress[0]
	}
	return types.Account(address)
}

func StarknetNodesParser(buf io.ReadCloser) ([]commonMonitoring.NodeConfig, error) {
	rawNodes := []StarknetNodeConfig{}
	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(&rawNodes); err != nil {
		return nil, fmt.Errorf("unable to unmarshal nodes config data: %w", err)
	}
	nodes := make([]commonMonitoring.NodeConfig, len(rawNodes))
	for i, rawNode := range rawNodes {
		nodes[i] = rawNode
	}
	return nodes, nil
}

func MakeStarknetNodeConfigs(in []commonMonitoring.NodeConfig) (out []StarknetNodeConfig, err error) {
	for i := range in {
		if in[i] == nil {
			return nil, fmt.Errorf("node config is nil")
		}

		cfg, ok := in[i].(StarknetNodeConfig)
		if !ok {
			return nil, fmt.Errorf("expected NodeConfig to be of type config.StarknetNodeConfig not %T - node name %s", in[i], in[i].GetName())
		}
		out = append(out, cfg)
	}
	return out, nil
}
