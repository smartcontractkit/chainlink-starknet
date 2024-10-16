package gauntlet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"context"
	g "github.com/smartcontractkit/gauntlet-plus-plus/sdks/go-gauntlet/client"
)

var (
	sgpp *StarknetGauntletPlusPlus
)

type Request struct {
	Input map[string]any 		 `json:"input"`
	Command string					 `json:"command"`
}

type StarknetGauntletPlusPlus struct {
	client				*g.ClientWithResponses
	gr      			*http.Response
	providers 		*[]g.Provider
}

func (sgpp *StarknetGauntletPlusPlus) BuildProviders(address string, rpcUrl string, privateKey string) (*[]g.Provider) {
	var input map[string]*interface{}
	input = make(map[string]*interface{})
	addressValue := interface{}(address)
	input["address"] = &addressValue
	AccountProvider := g.Provider{
		Name: "basic-address",
		Type: "@chainlink/gauntlet-starknet/lib/starknet.js/account",
		Input: input,
	}

	input = make(map[string]*interface{})
	privateKeyValue := interface{}(privateKey)
	debugValue := interface{}(true)
	input["privateKey"] = &privateKeyValue
	input["debug"] = &debugValue
	SignerProvider := g.Provider{
		Name: "basic-pk",
		Type: "@chainlink/gauntlet-starknet/lib/starknet.js/signer",
		Input: input,
	}

	input = make(map[string]*interface{})
	rpcUrlValue := interface{}(rpcUrl)
	input["url"] = &rpcUrlValue
	RpcProvider := g.Provider{
		Name: "basic-url",
		Type: "@chainlink/gauntlet-starknet/lib/starknet.js/provider",
		Input: input,
	}

	providers := []g.Provider{AccountProvider, SignerProvider, RpcProvider}

	return &providers
}

// New StarknetGauntletPlusPlus creates a default g++ client with responses
func NewStarknetGauntletPlusPlus(gauntletPPEndpoint string, rpcUrl string, address string, privateKey string) (*StarknetGauntletPlusPlus, error) {
	fmt.Println("rpcUrl: " + rpcUrl)
	fmt.Println("GPP URL: " + gauntletPPEndpoint)
	newClient, err := g.NewClientWithResponses(gauntletPPEndpoint)

	if err != nil {
		return nil, err
	}

	sgpp = &StarknetGauntletPlusPlus{
		client: newClient,
		gr: &http.Response{},
		providers: sgpp.BuildProviders(address, rpcUrl, privateKey),
	}

	return sgpp, nil
}

func (sgpp *StarknetGauntletPlusPlus) DeclareLinkTokenContract() (error) {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/token/link:declare",
		Input: inputMap,
	}

	var body g.PostExecuteJSONRequestBody
	var headers *g.PostExecuteParams

	body = *sgpp.BuildRequestBody(request)

	tmp,_ := json.Marshal(body)
	err := json.Unmarshal(tmp, &body)
	if err != nil {
		return err
	}

	// Show request body
	fmt.Println("Request body:" + string(tmp))

	// PostOperationWithResponse returns an already parsed Post Operation
	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, body)
	if err != nil {
		return err
	}

	fmt.Println("Response Body: " + string(response.Body))
	return err
}

func (sgpp *StarknetGauntletPlusPlus) DeployLinkTokenContract(address string) (string, error) {
	var contractAddress string
	input := make(map[string]interface{})
	minterValue := interface{}(address)
	ownerValue := interface{}(address)
	input["minter"] = &minterValue
	input["owner"] = &ownerValue
	request := Request{
		Command: "starknet/token/link:declare",
		Input: input,
	}

	var body g.PostExecuteJSONRequestBody
	var headers *g.PostExecuteParams

	body = *sgpp.BuildRequestBody(request)

	tmp,_ := json.Marshal(body)
	err := json.Unmarshal(tmp, &body)
	if err != nil {
		return "", nil
	}

	// Show request body
	fmt.Println(string(tmp))

	// PostOperationWithResponse returns an already parsed Post Operation
	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, body)
	if err != nil {
		return "", nil
	}

	report := response.JSON200

	fmt.Println(string(response.StatusCode()))
	if report.Output != nil {
		// Attempt to assert the Output as a map
		if outputMap, ok := (*report.Output).(map[string]interface{}); ok {
			if address, exists := outputMap["contractAddress"]; exists {
				fmt.Println("Address:", address)
			}
		} else {
			fmt.Println("Output is not of type map[string]interface{}")
		}
	} else {
		fmt.Println("Output is nil")
	}
	return contractAddress, nil
}

func (sgpp *StarknetGauntletPlusPlus) BuildRequestBody(request Request) (*g.PostExecuteJSONRequestBody) {
	body := g.PostExecuteJSONRequestBody{
		Config: &g.Config{
			Providers: *sgpp.providers,
			Datasources: []g.Datasource{},
		},
		Operation: g.Operation{
			Args: new(interface{}),
			Name: request.Command,
		},
	}

	*body.Operation.Args = request.Input

	return &body
}
