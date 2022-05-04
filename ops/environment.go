package ops

import "github.com/smartcontractkit/helmenv/environment"

func DefaultStarkNETEnv() *environment.Config {
	return &environment.Config{
		NamespacePrefix: "chainlink-starknet",
		Charts: environment.Charts{
			"geth": {Index: 1},
			"starknet": {
				Index: 1,
				Path:  "../../ops/charts/starknet",
				Values: map[string]interface{}{
					"real_node": false,
				},
			},
			"mockserver-config": {Index: 1},
			"mockserver":        {Index: 2},
			"chainlink": {
				Index: 2,
				Values: map[string]interface{}{
					"replicas": 5,
					"chainlink": map[string]interface{}{
						"image": map[string]interface{}{
							"image":   "795953128386.dkr.ecr.us-west-2.amazonaws.com/chainlink",
							"version": "latest.ff4e8e66be38bfb623a40589efd8668382af7cf1",
						},
					},
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
