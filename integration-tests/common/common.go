package common

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/chainlink"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver"
	mockservercfg "github.com/smartcontractkit/chainlink-env/pkg/helm/mockserver-cfg"
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

// 1. Deploy EVM nodes with OCRv2 config
// 2. Deploy contracts to EVM chain
// 3. Create bootstrap job spec for bootstrap node
// 4. Create job specs for child nodes
// 5. Create P2P keys

type Common struct {
	P2PPort             string
	ServiceKeyL1        string
	ServiceKeyL2        string
	ServiceKeyChainlink string
	ChainName           string
	ChainId             string
}

// CreateKeys Creates node keys and defines chain and nodes for each node
func (c *Common) CreateKeys(env *environment.Environment) ([]ctfClient.NodeKeysBundle, []*client.Chainlink, error) {
	chainlinkNodes, err := client.ConnectChainlinkNodes(env)
	if err != nil {
		return nil, nil, err
	}
	nKeys, err := ctfClient.CreateNodeKeysBundle(chainlinkNodes, c.ChainName)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range chainlinkNodes {
		_, _, err = n.CreateStarkNetChain(&client.StarkNetChainAttributes{
			Type:    c.ChainName,
			ChainID: c.ChainId,
			Config:  client.StarkNetChainConfig{},
		})
		if err != nil {
			return nil, nil, err
		}
		_, _, err = n.CreateStarkNetNode(&client.StarkNetNodeAttributes{
			Name:    c.ChainName,
			ChainID: c.ChainId,
			Url:     env.URLs[c.ServiceKeyL2][1],
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return nKeys, chainlinkNodes, nil
}

// CreateJobsForContract Creates and sets up the boostrap jobs as well as OCR jobs
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, observationSource string, juelsPerFeeCoinSource string, ocrControllerAddress string) error {
	// Define node[0] as bootstrap node
	cc.bootstrapPeers = []client.P2PData{
		{
			RemoteIP:   cc.chainlinkNodes[0].RemoteIP(),
			RemotePort: c.P2PPort,
			PeerID:     cc.nKeys[0].PeerID,
		},
	}

	// Defining relay config
	relayConfig := map[string]string{
		"nodeName": fmt.Sprintf("starknet-OCRv2-%s-%s", "node", uuid.NewV4().String()),
		"chainID":  c.ChainId,
	}

	// Setting up bootstrap node
	jobSpec := &client.OCR2TaskJobSpec{
		Name:                  fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.NewV4().String()),
		JobType:               "bootstrap",
		ContractID:            ocrControllerAddress,
		Relay:                 c.ChainName,
		RelayConfig:           relayConfig,
		ContractConfirmations: 1, // don't wait for confirmation on devnet
	}
	_, _, err := cc.chainlinkNodes[0].CreateJob(jobSpec)
	if err != nil {
		return err
	}

	// Setting up job specs
	for nIdx, n := range cc.chainlinkNodes {
		if nIdx == 0 {
			continue
		}
		_, err = n.CreateBridge(cc.bTypeAttr)
		if err != nil {
			return err
		}
		jobSpec := &client.OCR2TaskJobSpec{
			Name:                  fmt.Sprintf("starknet-OCRv2-%d-%s", nIdx, uuid.NewV4().String()),
			JobType:               "offchainreporting2",
			ContractID:            ocrControllerAddress,
			Relay:                 c.ChainName,
			RelayConfig:           relayConfig,
			PluginType:            "median",
			P2PV2Bootstrappers:    cc.bootstrapPeers,
			OCRKeyBundleID:        cc.nKeys[nIdx].OCR2Key.Data.ID,
			TransmitterID:         cc.nKeys[nIdx].TXKey.Data.ID,
			ObservationSource:     observationSource,
			JuelsPerFeeCoinSource: juelsPerFeeCoinSource,
			ContractConfirmations: 1, // don't wait for confirmation on devnet
		}
		_, _, err := n.CreateJob(jobSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Common) GetConfig() *Common {
	return c
}

func SetConfig(cfg *Common) *Common {
	return &Common{
		ServiceKeyChainlink: cfg.ServiceKeyChainlink,
		ServiceKeyL1:        cfg.ServiceKeyL1,
		ServiceKeyL2:        cfg.ServiceKeyL2,
		P2PPort:             cfg.P2PPort,
		ChainId:             cfg.ChainId,
		ChainName:           cfg.ChainName,
	}
}

func GetDefaultCoreConfig() map[string]interface{} {
	return map[string]interface{}{
		"STARKNET_ENABLED":            "true",
		"EVM_ENABLED":                 "false",
		"EVM_RPC_ENABLED":             "false",
		"CHAINLINK_DEV":               "false",
		"FEATURE_OFFCHAIN_REPORTING2": "true",
		"feature_offchain_reporting":  "false",
		"P2P_NETWORKING_STACK":        "V2",
		"P2PV2_LISTEN_ADDRESSES":      "0.0.0.0:6690",
		"P2PV2_DELTA_DIAL":            "5s",
		"P2PV2_DELTA_RECONCILE":       "5s",
		"p2p_listen_port":             "0",
	}
}

func GetDefaultEnvSetup(envConfig *environment.Config, clConfig map[string]interface{}) *environment.Environment {
	return environment.New(envConfig).
		// AddHelm(hardhat.New(nil)).
		AddHelm(devnet.New(nil)).
		AddHelm(mockservercfg.New(nil)).
		AddHelm(mockserver.New(nil)).
		AddHelm(chainlink.New(0, clConfig))
}
