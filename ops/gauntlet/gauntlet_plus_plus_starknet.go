package gauntlet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"context"
	"time"
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
	checkStatusValue := interface{}(true)
	input["url"] = &rpcUrlValue
	input["checkStatus"] = &checkStatusValue
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

func (sgpp *StarknetGauntletPlusPlus) TransferToken(tokenAddress string, to string, from string) (error) {
	inputMap := make(map[string]interface{})
	input := make(map[string]interface{})
	input["to"] = &to
	input["from"] = &from
	input["address"] = &tokenAddress
	request := Request{
		Command: "starknet/token/erc20:transfer",
		Input: inputMap,
	}

	var headers *g.PostExecuteParams

	body := sgpp.BuildRequestBody(request)

	tmp, _ := json.Marshal(body)

	// Show request body
	fmt.Println(string(tmp))

	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return err
	}

	fmt.Println("Response Body: " + string(response.Body))
	return nil
}

func (sgpp *StarknetGauntletPlusPlus) DeclareAccessControllerContract() (error) {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/data-feeds/access-controller@1.0.0:declare",
		Input: inputMap,
	}

	var headers *g.PostExecuteParams

	body := sgpp.BuildRequestBody(request)

	tmp, _ := json.Marshal(body)

	// Show request body
	fmt.Println(string(tmp))

	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return err
	}

	fmt.Println("Response Body: " + string(response.Body))
	return nil
}

func (sgpp *StarknetGauntletPlusPlus) DeployAccessControllerContract(address string) (string, error) {
	var contractAddress string
	input := make(map[string]interface{})
	constructorCalldata := make(map[string]interface{})
	ownerValue := interface{}(address)
	constructorCalldata["owner"] = &ownerValue
	input["constructorCalldata"] = &constructorCalldata
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
	contractAddress = sgpp.ExtractValueFromResponseBody(report, "contractAddress")

	return contractAddress, nil

}

func (sgpp *StarknetGauntletPlusPlus) DeclareLinkTokenContract() (error) {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/token/link:declare",
		Input: inputMap,
	}

	var headers *g.PostExecuteParams

	body := sgpp.BuildRequestBody(request)

	tmp, _ := json.Marshal(body)

	// Show request body
	fmt.Println(string(tmp))

	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return err
	}

	fmt.Println("Response Body: " + string(response.Body))
	time.Sleep(503024 * time.Minute)
	return nil
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
	contractAddress = sgpp.ExtractValueFromResponseBody(report, "contractAddress")

	return contractAddress, nil
}

func (sgpp *StarknetGauntletPlusPlus) ExtractValueFromResponseBody(report *g.Report, key string) string {
	if report.Output != nil {
		// Attempt to assert the Output as a map
		if outputMap, ok := (*report.Output).(map[string]interface{}); ok {
			if value, exists := outputMap[key]; exists {
				// Assert value to a string
				if strValue, ok := value.(string); ok {
					fmt.Println("Value:", strValue)
					return strValue
				} else {
					fmt.Println("Value is not of type string")
				}
			}
		} else {
			fmt.Println("Output is not of type map[string]interface{}")
		}
	} else {
		fmt.Println("Output is nil")
	}
	return ""
}

func (sgpp *StarknetGauntletPlusPlus) BuildRequestBody(request Request) (*g.PostExecuteJSONRequestBody) {
	var args any = request.Input

	body := g.PostExecuteJSONRequestBody{
		Config: &g.Config{
			Providers: *sgpp.providers,
			Datasources: []g.Datasource{},
		},
		Operation: g.Operation{
			Args: &args,
			Name: request.Command,
		},
	}

	return &body
}
