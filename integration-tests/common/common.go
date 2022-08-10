package common

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink-env/environment"
	ctfClient "github.com/smartcontractkit/chainlink-testing-framework/client"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
)

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
		_, _, err = n.CreateStarknetChain(&client.StarknetChainAttributes{
			Type:    c.ChainName,
			ChainID: c.ChainId,
			Config:  client.StarknetChainConfig{},
		})
		if err != nil {
			return nil, nil, err
		}
		_, _, err = n.CreateStarknetNode(&client.StarknetNodeAttributes{
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
func (c *Common) CreateJobsForContract(cc *ChainlinkClient, juelsPerFeeCoinSource string, ocrControllerAddress string) error {
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
		Name:        fmt.Sprintf("starknet-OCRv2-%s-%s", "bootstrap", uuid.NewV4().String()),
		JobType:     "bootstrap",
		ContractID:  ocrControllerAddress,
		Relay:       c.ChainName,
		RelayConfig: relayConfig,
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
			ObservationSource:     client.ObservationSourceSpecBridge(*cc.bTypeAttr),
			JuelsPerFeeCoinSource: juelsPerFeeCoinSource,
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
