package infra_deployments_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/common"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/chainlink/integration-tests/client"
	"time"
)

var (
	observationSource = `
			val [type="bridge" name="bridge-mockserver" requestData=<{"data": {"from":"LINK","to":"USD"}}>]
			parse [type="jsonparse" path="result"]
			val -> parse
			`
	juelsPerFeeCoinSource = `"""
			sum  [type="sum" values=<[451000]> ]
			sum
			"""
			`
)

func createKeys() ([]*client.Chainlink, error) {
	urls := [][]string{}
	var clients []*client.Chainlink

	for _, url := range urls {
		c, err := client.NewChainlink(&client.ChainlinkConfig{
			URL:      url[0],
			Email:    url[1],
			Password: url[2],
			RemoteIP: url[0],
		})
		if err != nil {
			return nil, err
		}
		key, _ := c.MustReadP2PKeys()
		if key == nil {
			_, _, err = c.CreateP2PKey()
			Expect(err).ShouldNot(HaveOccurred())
		}
		clients = append(clients, c)
	}

	return clients, nil
}

var _ = Describe("@infra", func() {
	It("works", func() {
		var err error
		lggr := logger.Nop()
		rpcRequestTimeout := time.Second * 300
		t := &common.Test{}
		t.Common = common.New()
		t.Common.Default()
		t.Cc = &common.ChainlinkClient{}
		//t.Common.P2PPort = "6690"
		t.Devnet = t.Devnet.NewStarkNetDevnetClient(t.Common.L2RPCUrl, "")
		t.Cc.ChainlinkNodes, err = createKeys()
		t.Cc.NKeys, _, err = client.CreateNodeKeysBundle(t.Cc.ChainlinkNodes, t.Common.ChainName, t.Common.ChainId)
		Expect(err).ShouldNot(HaveOccurred())
		for _, n := range t.Cc.ChainlinkNodes {
			_, _, err = n.CreateStarkNetChain(&client.StarkNetChainAttributes{
				Type:    t.Common.ChainName,
				ChainID: t.Common.ChainId,
				Config:  client.StarkNetChainConfig{},
			})
			Expect(err).ShouldNot(HaveOccurred())
			_, _, err = n.CreateStarkNetNode(&client.StarkNetNodeAttributes{
				Name:    t.Common.ChainName,
				ChainID: t.Common.ChainId,
				Url:     "https://alpha4-2.starknet.io",
			})
			Expect(err).ShouldNot(HaveOccurred())
		}
		t.Starknet, err = starknet.NewClient(t.Common.ChainId, t.Common.L2RPCUrl, lggr, &rpcRequestTimeout)
		t.Common.Testnet = true
		t.Common.L2RPCUrl = "https://alpha4-2.starknet.io"
		//t.Sg, err = gauntlet.NewStarknetGauntlet("../../")
		//Expect(err).ShouldNot(HaveOccurred(), "Could not get a new gauntlet struct")
		//
		//err = t.Sg.SetupNetwork(t.Common.L2RPCUrl)
		//Expect(err).ShouldNot(HaveOccurred(), "Setting up gauntlet network should not fail")
		//err = t.DeployGauntlet(-100000000000, 100000000000, 9, "auto", 1, 1)
		//Expect(err).ShouldNot(HaveOccurred(), "Deploying contracts should not fail")
		t.SetBridgeTypeAttrs(&client.BridgeTypeAttributes{
			Name: "bridge-mockserver",
			URL:  "https://adapters.main.stage.cldev.sh/coinmetrics",
		})
		t.Common.P2PPort = "6690"
		err = t.Common.CreateJobsForContract(t.Cc, observationSource, juelsPerFeeCoinSource, t.OCRAddr)
	})

})
