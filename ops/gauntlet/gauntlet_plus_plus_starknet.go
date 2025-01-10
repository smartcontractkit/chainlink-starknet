package gauntlet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	g "github.com/smartcontractkit/gauntlet-plus-plus/sdks/go-gauntlet/client"
)

var (
	sgpp *StarknetGauntletPlusPlus
)

type Request struct {
	Input   map[string]any `json:"input"`
	Command string         `json:"command"`
}

type StarknetGauntletPlusPlus struct {
	client    *g.ClientWithResponses
	gr        *http.Response
	providers *[]g.Provider
}

func (sgpp *StarknetGauntletPlusPlus) BuildProviders(address string, rpcURL string, privateKey string) *[]g.Provider {
	accountProviderInput := map[string]interface{}{
		"address": address,
	}
	AccountProvider := g.Provider{
		Name:  "basic-address",
		Type:  "@chainlink/gauntlet-starknet/lib/starknet.js/account",
		Input: toPointerMap(accountProviderInput),
	}

	signerProviderInput := map[string]interface{}{
		"privateKey": privateKey,
		"debug":      true,
	}
	SignerProvider := g.Provider{
		Name:  "basic-pk",
		Type:  "@chainlink/gauntlet-starknet/lib/starknet.js/signer",
		Input: toPointerMap(signerProviderInput),
	}

	providerInput := map[string]interface{}{
		"url":         rpcURL,
		"checkStatus": false,
	}
	RPCProvider := g.Provider{
		Name:  "basic-url",
		Type:  "@chainlink/gauntlet-starknet/lib/starknet.js/provider",
		Input: toPointerMap(providerInput),
	}

	providers := []g.Provider{AccountProvider, SignerProvider, RPCProvider}

	return &providers
}

// New StarknetGauntletPlusPlus creates a default g++ client with responses
func NewStarknetGauntletPlusPlus(gauntletPPEndpoint string, rpcURL string, address string, privateKey string) (*StarknetGauntletPlusPlus, error) {
	log.Info().Str("Creating G++ Client with Endpoint: ", gauntletPPEndpoint).Msg("Gauntlet++")
	log.Info().Str("Connecting G++ Client to RPC URL: ", rpcURL).Msg("Gauntlet++")
	newClient, err := g.NewClientWithResponses(gauntletPPEndpoint)

	if err != nil {
		return nil, err
	}

	sgpp = &StarknetGauntletPlusPlus{
		client:    newClient,
		gr:        &http.Response{},
		providers: sgpp.BuildProviders(address, rpcURL, privateKey),
	}

	return sgpp, nil
}

func (sgpp *StarknetGauntletPlusPlus) ExtractValueFromResponseBody(report g.Report, key string) (string, error) {
	if report.Output != nil {
		// Log the raw content of Output
		_, err := json.Marshal(report.Output)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal report.Output")
			return "", err
		}

		// Attempt to assert the Output as a map
		if outputMap, ok := (*report.Output).(map[string]interface{}); ok {
			log.Info().Interface("Report Response: ", outputMap).Msg("Gauntlet++")

			if value, exists := outputMap[key]; exists {
				// Assert value to a string
				if strValue, ok := value.(string); ok {
					return strValue, nil
				}
				err := fmt.Errorf("parsed Value is not of type string")
				return "", err
			}
		} else {
			// Log a message if it’s not a map
			log.Warn().Msg("Report.Output is not a map[string]interface{}")
		}
	}
	return "", nil
}

func (sgpp *StarknetGauntletPlusPlus) BuildRequestBody(request Request) *g.PostExecuteJSONRequestBody {
	var args any = request.Input

	body := g.PostExecuteJSONRequestBody{
		Config: &g.Config{
			Providers:   *sgpp.providers,
			Datasources: []g.Datasource{},
		},
		Operation: g.Operation{
			Args: &args,
			Name: request.Command,
		},
	}

	return &body
}

func (sgpp *StarknetGauntletPlusPlus) execute(request *Request) error {
	report, err := sgpp.executeReturnsReport(request)

	if err != nil {
		return err // Handle post execution error
	}

	if report.Output != nil {
		_, err := json.Marshal(report.Output)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal report.Output")
			return err
		}
		err = processReport(&report)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process Op report")
			return err
		}
	}
	return nil
}

func (sgpp *StarknetGauntletPlusPlus) executeReturnsReport(request *Request) (g.Report, error) {
	body := sgpp.BuildRequestBody(*request)

	tmp, err := json.Marshal(body)
	if err != nil {
		return g.Report{}, err // Handle marshaling error
	}

	// Show request body
	log.Info().Str("Request Body:", string(tmp)).Msg("Gauntlet++")

	// Make the API call
	headers := &g.PostExecuteParams{}

	response, err := sgpp.client.PostExecuteWithResponse(context.Background(), headers, *body)
	if err != nil {
		return g.Report{}, err // Handle post execution error
	}

	// Log the response body
	responseJSON, err := json.Marshal(response.JSON200) // Marshal the JSON200 field to JSON string
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response body")
		return g.Report{}, err
	}
	if response.JSON200 == nil || response.JSON200.Id == "" || response == nil {
		time.Sleep(20 * time.Minute)
	}

	// Log the full response JSON
	log.Info().Str("Response Body:", string(responseJSON)).Msg("Gauntlet++")

	return *response.JSON200, nil
}

func (sgpp *StarknetGauntletPlusPlus) executeDeploy(request *Request) (string, error) {
	report, err := sgpp.executeReturnsReport(request)

	if err != nil {
		return "", err // Handle post execution error
	}

	contractAddress, err := sgpp.ExtractValueFromResponseBody(report, "contractAddress")
	if err != nil {
		log.Err(err).Str("G++ Request returned with err", err.Error()).Msg("Gauntlet++")
		return "", err
	}

	if contractAddress == "" {
		log.Err(err).Str("G++ Deploy Requets returned with empty contractAddress", err.Error()).Msg("Gauntlet++")
		return "", err
	}

	log.Info().Str("Contract Address Response: ", contractAddress).Msg("Gauntlet++")
	return contractAddress, nil
}

func (sgpp *StarknetGauntletPlusPlus) TransferToken(tokenAddress string, to string, amount string) error {
	inputMap := map[string]interface{}{
		"address": tokenAddress,
		"to":      to,
		"amount":  amount,
	}

	request := Request{
		Command: "starknet/token/erc20:transfer",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareOCR2Controllercontract() error {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/data-feeds/aggregator@1.0.0:declare",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeployOCR2ControllerContract(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string,
	linkTokenAddress string, address string, accessControllerAddress string) (string, error) {
	// Delare Contract First
	err := sgpp.DeclareOCR2Controllercontract()
	if err != nil {
		return "", err
	}

	constructorCalldata := map[string]interface{}{
		"owner":                   address,
		"link":                    linkTokenAddress,
		"minAnswer":               minSubmissionValue,
		"maxAnswer":               maxSubmissionValue,
		"billingAccessController": accessControllerAddress,
		"decimals":                decimals,
		"description":             "USDT/LINK",
	}
	inputMap := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	request := Request{
		Command: "starknet/data-feeds/aggregator@1.0.0:deploy",
		Input:   inputMap,
	}

	return sgpp.executeDeploy(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareOCR2ControllerProxyContract() error {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/data-feeds/aggregator-proxy@1.0.0:declare",
		Input:   inputMap,
	}
	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeployOCR2ControllerProxyContract(address string, controllerContractAddress string) (string, error) {
	// Declare Contract First
	err := sgpp.DeclareOCR2ControllerProxyContract()
	if err != nil {
		return "", err
	}

	constructorCalldata := map[string]interface{}{
		"owner":   address,
		"address": controllerContractAddress,
	}
	inputMap := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	request := Request{
		Command: "starknet/data-feeds/aggregator-proxy@1.0.0:deploy",
		Input:   inputMap,
	}

	return sgpp.executeDeploy(&request)
}

func (sgpp *StarknetGauntletPlusPlus) AddAccess(aggregatorAddress string, grantAddress string) error {
	inputMap := map[string]interface{}{
		"address":      aggregatorAddress,
		"grantAddress": grantAddress,
	}

	request := Request{
		Command: "starknet/data-feeds/access-controller@1.0.0:add-access",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareAccessControllerContract() error {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/data-feeds/access-controller@1.0.0:declare",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeployAccessControllerContract(address string) (string, error) {
	constructorCalldata := map[string]interface{}{
		"owner": address,
	}
	inputMap := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	request := Request{
		Command: "starknet/data-feeds/access-controller@1.0.0:deploy",
		Input:   inputMap,
	}
	return sgpp.executeDeploy(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareLinkTokenContract() error {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/token/link:declare",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeployLinkTokenContract(address string) (string, error) {
	// Declare token first
	err := sgpp.DeclareLinkTokenContract()

	if err != nil {
		return "", err
	}

	constructorCalldata := map[string]interface{}{
		"minter": address,
		"owner":  address,
	}
	inputMap := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	request := Request{
		Command: "starknet/token/link:deploy",
		Input:   inputMap,
	}

	return sgpp.executeDeploy(&request)
}

func (sgpp *StarknetGauntletPlusPlus) SetConfigDetails(cfg string, ocrAddress string) (g.Report, error) {
	txArgs := make(map[string]interface{})
	err := json.Unmarshal([]byte(cfg), &txArgs)
	if err != nil {
		// Handle the error appropriately (return, log, etc.)
		return g.Report{}, nil
	}
	inputMap := map[string]interface{}{
		"address": ocrAddress,
		"txArgs":  &txArgs,
	}
	request := Request{
		Command: "starknet/data-feeds/aggregator@1.0.0:set-config",
		Input:   inputMap,
	}
	test, testerr := sgpp.executeReturnsReport(&request)
	return test, testerr
}

func (sgpp *StarknetGauntletPlusPlus) SetOCRBilling(observationPaymentGjuels int64, transmissionPaymentGjuels int64, ocrAddress string) (g.Report, error) {
	txArgs := map[string]interface{}{
		"transmissionPaymentGjuels": transmissionPaymentGjuels,
		"observationPaymentGjuels":  observationPaymentGjuels,
		"gasPerSignature":           0,
		"gasBase":                   0,
	}
	inputMap := map[string]interface{}{
		"address": ocrAddress,
		"txArgs":  &txArgs,
	}

	request := Request{
		Command: "starknet/data-feeds/aggregator@1.0.0:set-billing",
		Input:   inputMap,
	}

	return sgpp.executeReturnsReport(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeclareOzAccount() error {
	inputMap := make(map[string]interface{})
	request := Request{
		Command: "starknet/chain/open-zeppelin:declare",
		Input:   inputMap,
	}

	return sgpp.execute(&request)
}

func (sgpp *StarknetGauntletPlusPlus) DeployOzAccount(publicKey string) (string, error) {
	constructorCalldata := map[string]interface{}{
		"publicKey": publicKey,
	}
	inputMap := map[string]interface{}{
		"constructorCalldata": &constructorCalldata,
	}

	request := Request{
		Command: "starknet/chain/open-zeppelin:deploy",
		Input:   inputMap,
	}

	return sgpp.executeDeploy(&request)
}

func toPointerMap(input map[string]interface{}) map[string]*interface{} {
	result := make(map[string]*interface{})
	for k, v := range input {
		// Create a new variable to hold the value
		valueCopy := v
		// Store the pointer to the new variable
		result[k] = &valueCopy
	}
	return result
}

func processReport(report *g.Report) error {
	// Ensure Output is a map
	outputMap, ok := (*report.Output).(map[string]interface{})
	if !ok {
		log.Warn().Msg("Report.Output is not a map[string]interface{}")
		return fmt.Errorf("Report.Output is not a map")
	}

	log.Info().Interface("Report Response: ", outputMap).Msg("Gauntlet++")

	// Access the 'receipt' field
	receiptMap, err := getReceiptMap(outputMap)
	if err != nil {
		return err
	}

	// Check 'execution_status' inside the 'receipt' field
	return checkExecutionStatus(receiptMap)
}

// Helper function to extract the receipt map
func getReceiptMap(outputMap map[string]interface{}) (map[string]interface{}, error) {
	output, exists := outputMap["receipt"]
	if !exists {
		return nil, fmt.Errorf("receipt does not exist")
	}

	receiptMap, ok := output.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("receipt is not a map")
	}

	log.Info().Interface("Receipt Map: ", receiptMap).Msg("Gauntlet++")
	return receiptMap, nil
}

// Helper function to check the execution status
func checkExecutionStatus(receiptMap map[string]interface{}) error {
	executionStatus, exists := receiptMap["execution_status"]
	if !exists {
		return fmt.Errorf("execution_status does not exist")
	}

	strExecutionStatus, ok := executionStatus.(string)
	if !ok {
		return fmt.Errorf("execution_status is not a string")
	}

	if strExecutionStatus != "SUCCEEDED" {
		return fmt.Errorf("Op was not successful")
	}

	return nil
}
