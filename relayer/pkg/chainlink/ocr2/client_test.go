package ocr2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

const BLOCK_OUTPUT = `{"block_hash": "0x2", "block_number": 2, "parent_block_hash": "0x1", "sequencer_address": "0x4bbfb0d1aab5bf33eec5ada3a1040c41ed902a1eeb38c78a753d6f6359f1666", "transactions": [{"entry_point_type": "EXTERNAL", "transaction_hash": "0x2bb11daf4649bbadd51eb33a6890abb869e55c9d65f424e7ef74d34b543b914", "max_fee": "0x1713081b1e9400", "signature": ["0x2dd6147385d73e6a1b278863aa8be6b3c545a81810406807862c15aaae459b8", "0x1d072fdd6280d74c1e0788af5cc0f0169323167f04162a15660afe2e32425a6"], "calldata": ["0x1", "0xd43963a4e875a361f5d164b2e70953598eb4f45fde86924082d51b4d78e489", "0x2f09d463867ebc70075969cdd4d3c0643f9b69dbc84bf9bf96923a52dc3c8d8", "0x0", "0x22", "0x22", "0x4", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737", "0x1", "0x0", "0x2", "0x15", "0x263", "0x880a0d9e61d1080d88ee16f1880bcc1960b2080cab5ee01288090dfc04a30", "0x53a0201024220af400004fa5d02cd5170b5261032e71f2847ead36159cf8d", "0xee68affc3c8520904220af400004fa5d02cd5170b5261032e71f2847ead361", "0x59cf8dee68affc3c8520914220af400004fa5d02cd5170b5261032e71f2847", "0xead36159cf8dee68affc3c8520924220af400004fa5d02cd5170b5261032e7", "0x1f2847ead36159cf8dee68affc3c8520934a42307830346363316266613939", "0x65323832653433346165663238313563613137333337613932336364326336", "0x31636630633764653562333236643761383630333733304a42307830346363", "0x31626661393965323832653433346165663238313563613137333337613932", "0x33636432633631636630633764653562333236643761383630333733314a42", "0x30783034636331626661393965323832653433346165663238313563613137", "0x33333761393233636432633631636630633764653562333236643761383630", "0x333733324a4230783034636331626661393965323832653433346165663238", "0x31356361313733333761393233636432633631636630633764653562333236", "0x643761383630333733335200608094ebdc03688084af5f708084af5f788084", "0xaf5f82018c010a202ac49e648a1f84da5a143eeab68c8402c65a1567e63971", "0x7f5732d5e6310c2c761220a6c1ae85186dc981dc61cd14d7511ee5ab70258a", "0x10ac4e03e4d4991761b2c0a61a1090696dc7afed7f61a26887e78e683a1c1a", "0x10a29e5fa535f2edea7afa9acb4fd349b31a10d1b88713982955d79fa0e422", "0x685a748b1a10a07e0118cc38a71d2a9d60bf52938b4a", "0x0"], "contract_address": "0x7e00d496e324876bbc8531f2d9a82bf154d1a04a50218ee74cdd372f75a551a", "entry_point_selector": "0x15d40a3d6ca2ac30f4031e42be28da9b056fef9bb7357ac5e85627ee876e5ad", "type": "INVOKE_FUNCTION"}], "starknet_version": "0.9.1", "transaction_receipts": [{"transaction_hash": "0x2bb11daf4649bbadd51eb33a6890abb869e55c9d65f424e7ef74d34b543b914", "actual_fee": "0xf62056769b800", "transaction_index": 0, "l2_to_l1_messages": [], "execution_resources": {"builtin_instance_counter": {"pedersen_builtin": 69, "range_check_builtin": 97, "ecdsa_builtin": 1, "bitwise_builtin": 1, "output_builtin": 0}, "n_steps": 3432, "n_memory_holes": 326}, "events": [{"from_address": "0xd43963a4e875a361f5d164b2e70953598eb4f45fde86924082d51b4d78e489", "keys": ["0x9a144bf4a6a8fd083c93211e163e59221578efcc86b93f8c97c620e7b9608a"], "data": ["0x0", "0x4b791b801cf0d7b6a2f9e59daf15ec2dd7d9cdc3bc5e037bada9c86e4821c", "0x1", "0x4", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733", "0x4cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737", "0x1", "0x3", "0x1", "0x0", "0xf4240", "0x2", "0x15", "0x263", "0x880a0d9e61d1080d88ee16f1880bcc1960b2080cab5ee01288090dfc04a30", "0x53a0201024220af400004fa5d02cd5170b5261032e71f2847ead36159cf8d", "0xee68affc3c8520904220af400004fa5d02cd5170b5261032e71f2847ead361", "0x59cf8dee68affc3c8520914220af400004fa5d02cd5170b5261032e71f2847", "0xead36159cf8dee68affc3c8520924220af400004fa5d02cd5170b5261032e7", "0x1f2847ead36159cf8dee68affc3c8520934a42307830346363316266613939", "0x65323832653433346165663238313563613137333337613932336364326336", "0x31636630633764653562333236643761383630333733304a42307830346363", "0x31626661393965323832653433346165663238313563613137333337613932", "0x33636432633631636630633764653562333236643761383630333733314a42", "0x30783034636331626661393965323832653433346165663238313563613137", "0x33333761393233636432633631636630633764653562333236643761383630", "0x333733324a4230783034636331626661393965323832653433346165663238", "0x31356361313733333761393233636432633631636630633764653562333236", "0x643761383630333733335200608094ebdc03688084af5f708084af5f788084", "0xaf5f82018c010a202ac49e648a1f84da5a143eeab68c8402c65a1567e63971", "0x7f5732d5e6310c2c761220a6c1ae85186dc981dc61cd14d7511ee5ab70258a", "0x10ac4e03e4d4991761b2c0a61a1090696dc7afed7f61a26887e78e683a1c1a", "0x10a29e5fa535f2edea7afa9acb4fd349b31a10d1b88713982955d79fa0e422", "0x685a748b1a10a07e0118cc38a71d2a9d60bf52938b4a"]}]}], "gas_price": "0x174876e800", "state_root": "0000000000000000000000000000000000000000000000000000000000000000", "status": "ACCEPTED_ON_L2", "timestamp": 1664483024}`
const ocr2ContractAddress = "0xd43963a4e875a361f5d164b2e70953598eb4f45fde86924082d51b4d78e489" // matches BLOCK_OUTPUT event

func TestOCR2Client(t *testing.T) {
	chainID := gateway.GOERLI_ID
	lggr := logger.Test(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := ioutil.ReadAll(r.Body)
		fmt.Println(r.RequestURI, r.URL, string(req))

		var out []byte

		switch {
		case r.RequestURI == "/feeder_gateway/call_contract":
			type Request struct {
				Selector string `json:"entry_point_selector"`
			}

			reqdata := Request{}
			require.NoError(t, json.Unmarshal(req, &reqdata))

			switch {
			case caigo.BigToHex(caigo.GetSelectorFromName("billing")) == reqdata.Selector:
				// billing response
				out = []byte(`{"result":["0x0","0x0","0x0","0x0"]}`)
			case caigo.BigToHex(caigo.GetSelectorFromName("latest_config_details")) == reqdata.Selector:
				// latest config details response
				out = []byte(`{"result":["0x1","0x2","0x4b791b801cf0d7b6a2f9e59daf15ec2dd7d9cdc3bc5e037bada9c86e4821c"]}`)
			case caigo.BigToHex(caigo.GetSelectorFromName("latest_transmission_details")) == reqdata.Selector:
				// latest transmission details response
				out = []byte(`{"result":["0x4cfc96325fa7d72e4854420e2d7b0abda72de17d45e4c3c0d9f626016d669","0x0","0x0","0x0"]}`)
			default:
				require.False(t, true, "unsupported contract method")
			}
		case strings.Contains(r.RequestURI, "/feeder_gateway/get_block"):
			out = []byte(BLOCK_OUTPUT)
		default:
			require.False(t, true, "unsupported endpoint")
		}

		_, err := w.Write(out)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	url := mockServer.URL
	duration := 10 * time.Second
	reader, err := starknet.NewClient(chainID, url, lggr, &duration)
	require.NoError(t, err)
	client, err := NewClient(reader, lggr)
	assert.NoError(t, err)

	t.Run("get billing details", func(t *testing.T) {
		billing, err := client.BillingDetails(context.Background(), ocr2ContractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", billing)
	})

	t.Run("get latest config details", func(t *testing.T) {
		details, err := client.LatestConfigDetails(context.Background(), ocr2ContractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", details)

		config, err := client.ConfigFromEventAt(context.Background(), ocr2ContractAddress, details.Block)
		require.NoError(t, err)
		fmt.Printf("%+v\n", config)
	})

	t.Run("get latest transmission details", func(t *testing.T) {
		transmissions, err := client.LatestTransmissionDetails(context.Background(), ocr2ContractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", transmissions)
	})

	t.Run("get latest round data", func(t *testing.T) {
		round, err := client.LatestRoundData(context.Background(), ocr2ContractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", round)
	})

	t.Run("get link available for payment", func(t *testing.T) {
		available, err := client.LinkAvailableForPayment(context.Background(), ocr2ContractAddress)
		require.NoError(t, err)
		fmt.Printf("%+v\n", available)
	})

	t.Run("get latest transmission", func(t *testing.T) {
		round, err := client.LatestRoundData(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", round)

		transmissions, err := client.NewTransmissionsFromEventsAt(context.Background(), ocr2ContractAddress, round.BlockNumber)
		assert.NoError(t, err)
		fmt.Printf("%+v\n", transmissions)
	})
}
