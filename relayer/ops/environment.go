package ops

import (
	"github.com/smartcontractkit/chainlink-testing-framework/config"
	"github.com/smartcontractkit/helmenv/environment"
)

func DefaultStarkNETEnv() *environment.Config {
	return &environment.Config{
		NamespacePrefix: "chainlink-starknet",
		Charts: environment.Charts{
			"geth": {Index: 1},
			"starknet": {
				Index: 1,
				Path:  "../../relayer/ops/charts/starknet",
				Values: map[string]interface{}{
					"real_node": false,
				},
			},
			"mockserver-config": {Index: 1},
			"mockserver":        {Index: 2},
			"chainlink": {
				Index: 2,
				Values: map[string]interface{}{
					"replicas":  5,
					"chainlink": config.ChainlinkVals(),
					"env": map[string]interface{}{
						"EVM_ENABLED":                 "true",
						"EVM_RPC_ENABLED":             "true",
						"eth_url":                     "ws://geth:8546",
						"eth_http_url":                "http://geth:8544",
						"feature_external_initiators": "true",
						"FEATURE_CCIP":                "true",
						"FEATURE_OFFCHAIN_REPORTING2": "true",
						"P2P_NETWORKING_STACK":        "V2",
						"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:6690",
						"P2PV2_DELTA_DIAL":            "5s",
						"P2PV2_DELTA_RECONCILE":       "5s",
						"p2p_listen_port":             "0",
					},
				},
			},
		},
	}
}
