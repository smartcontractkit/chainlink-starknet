package gauntlet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"context"
	"github.com/rs/zerolog/log"
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

func (sgpp *StarknetGauntletPlusPlus) ExtractValueFromResponseBody(report g.Report, key string) string {
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

func (sgpp *StarknetGauntletPlusPlus) executeRequest(command string, inputMap map[string]interface{}) error {
	request := Request{
		Command: command,
		Input:   inputMap,
	}

	body := sgpp.BuildRequestBody(request)

	tmp, err := json.Marshal(body)
	if err != nil {
		return err // Handle marshaling error
	}

	// Show request body
	log.Info().Str("Request Body: ", string(tmp)).Msg("Gauntlet++")

	headers := &g.PostExecuteParams{}
	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return err // Handle post execution error
	}

	// Show Response Status
	log.Info().Str("Response Status:", string(response.Status())).Msg("Gauntlet++")
	return nil
}

func (sgpp *StarknetGauntletPlusPlus) executeRequestReturnsReport(command string, inputMap map[string]interface{}) (g.Report, error) {
	request := Request{
		Command: command,
		Input:   inputMap,
	}

	body := sgpp.BuildRequestBody(request)

	tmp, err := json.Marshal(body)
	if err != nil {
		return g.Report{}, err // Handle marshaling error
	}

	// Show request body
	log.Info().Str("Request Body: ", string(tmp)).Msg("Gauntlet++")

	headers := &g.PostExecuteParams{}
	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return g.Report{}, err // Handle post execution error
	}

	return *response.JSON200, nil
}

func (sgpp *StarknetGauntletPlusPlus) executeDeployRequest(command string, inputMap map[string]interface{}) (string, error) {
	report, err := sgpp.executeRequestReturnsReport(command, inputMap)

	if err != nil {
		return "", err // Handle post execution error
	}
	contractAddress := sgpp.ExtractValueFromResponseBody(report, "contractAddress")

	return contractAddress, nil
}

func (sgpp *StarknetGauntletPlusPlus) TransferToken(tokenAddress string, to string, from string) (error) {
	inputMap := map[string]interface{}{
		"to":     to,
		"from": from,
		"address": tokenAddress,
	}

	return sgpp.executeRequest("starknet/token/erc20:transfer", inputMap)
}

func (sgpp *StarknetGauntletPlusPlus) DeployOCR2ControllerContract(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, 
	linkTokenAddress string, address string, accessControllerAddress string) (string, error) {		
		constructorCalldata := map[string]interface{}{
			"owner": address,
			"link": linkTokenAddress,
			"minAnswer": minSubmissionValue,
			"maxAnswer": maxSubmissionValue,
			"billingAccessController": accessControllerAddress,
			"decimals": decimals,
			"description": "USDT/LINK",

		}
		input := map[string]interface{}{
			"constructorCalldata": &constructorCalldata,
		}
	
		return sgpp.executeDeployRequest("starknet/data-feeds/aggregator@1.0.0:deploy", input)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareOCR2Controllercontract() (error) {
	inputMap := make(map[string]interface{})

	return sgpp.executeRequest("starknet/data-feeds/aggregator@1.0.0:declare", inputMap)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareOCR2ControllerProxyContract() (error) {
	inputMap := make(map[string]interface{})
	
	return sgpp.executeRequest("starknet/data-feeds/aggregator-proxy@1.0.0:declare", inputMap)
}

func (sgpp *StarknetGauntletPlusPlus) DeployOCR2ControllerProxyContract(address string, controllerContractAddress string) (string, error) {		
		constructorCalldata := map[string]interface{}{
			"owner": address,
			"address": controllerContractAddress,
		}
		input := map[string]interface{} {
			"constructorCalldata": &constructorCalldata,
		}
	
		return sgpp.executeDeployRequest("starknet/data-feeds/aggregator-proxy@1.0.0:deploy", input)
}

func (sgpp *StarknetGauntletPlusPlus) AddAccess(aggregatorAddress string, grantAddress string)  (error) {
	inputMap := map[string]interface{} {
		"address": aggregatorAddress,
		"grantAddress": grantAddress,
	}

	return sgpp.executeRequest("starknet/data-feeds/access-controller@1.0.0:add-access", inputMap)
}



func (sgpp *StarknetGauntletPlusPlus) DeclareAccessControllerContract() (error) {
	inputMap := make(map[string]interface{})

	return sgpp.executeRequest("starknet/data-feeds/access-controller@1.0.0:declare", inputMap)
}

func (sgpp *StarknetGauntletPlusPlus) DeployAccessControllerContract(address string) (string, error) {	
	constructorCalldata := map[string]interface{} {
		"owner": address,
	}
	input := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	return sgpp.executeDeployRequest("starknet/token/link:declare", input)

}

func (sgpp *StarknetGauntletPlusPlus) DeclareLinkTokenContract() (error) {
	inputMap := make(map[string]interface{})

	return sgpp.executeRequest("starknet/token/link:declare", inputMap)
}

func (sgpp *StarknetGauntletPlusPlus) DeployLinkTokenContract(address string) (string, error) {
	input := map[string]interface{}{
		"minter": address,
		"owner": address,
	}

	return sgpp.executeDeployRequest("starknet/token/link:deploy", input)
}

func (sgpp *StarknetGauntletPlusPlus) SetConfigDetails(cfg string, ocrAddress string) (g.Report, error) {
	txArgs := make(map[string]interface{})
	err := json.Unmarshal([]byte(cfg), &txArgs)
	if err != nil {
		// Handle the error appropriately (return, log, etc.)
		return g.Report{}, nil 
	}
	input := map[string]interface{}{
		"address": ocrAddress,
		"txArgs": &txArgs,
	}

	return sgpp.executeRequestReturnsReport("starknet/data-feeds/aggregator@1.0.0:set-config", input)

}

func (sgpp *StarknetGauntletPlusPlus) SetOCRBilling(observationPaymentGjuels int64, transmissionPaymentGjuels int64, ocrAddress string) (g.Report, error) {
	txArgs := map[string]interface{} {
		"transmissionPaymentGjuels": transmissionPaymentGjuels,
		"observationPaymentGjuels": observationPaymentGjuels,
		"gasPerSignature": "0",
		"gasBase": "0",
	}
	input := map[string]interface{} {
		"address": ocrAddress,
		"txArgs": &txArgs,
	}

	return sgpp.executeRequestReturnsReport("starknet/data-feeds/aggregator@1.0.0:set-billing", input)
}
