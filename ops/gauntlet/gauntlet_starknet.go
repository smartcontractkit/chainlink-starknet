package gauntlet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/smartcontractkit/chainlink-testing-framework/gauntlet"
)

var (
	sg *StarknetGauntlet
)

type StarknetGauntlet struct {
	dir     string
	G       *gauntlet.Gauntlet
	gr      *GauntletResponse
	options *gauntlet.ExecCommandOptions
}

// GauntletResponse Default response output for starknet gauntlet commands
type GauntletResponse struct {
	Responses []struct {
		Tx struct {
			Hash    string `json:"hash"`
			Address string `json:"address"`
			Status  string `json:"status"`

			Tx struct {
				Address         string   `json:"address"`
				Code            string   `json:"code"`
				Result          []string `json:"result"`
				TransactionHash string   `json:"transaction_hash"`
			} `json:"tx"`
		} `json:"tx"`
		Contract string `json:"contract"`
	} `json:"responses"`
}

// NewStarknetGauntlet Creates a default gauntlet config
func NewStarknetGauntlet(workingDir string) (*StarknetGauntlet, error) {
	g, err := gauntlet.NewGauntlet()
	g.SetWorkingDir(workingDir)
	if err != nil {
		return nil, err
	}
	sg = &StarknetGauntlet{
		dir: workingDir,
		G:   g,
		gr:  &GauntletResponse{},
		options: &gauntlet.ExecCommandOptions{
			ErrHandling:       []string{},
			CheckErrorsInRead: true,
		},
	}
	return sg, nil
}

// FetchGauntletJsonOutput Parse gauntlet json response that is generated after yarn gauntlet command execution
func (sg *StarknetGauntlet) FetchGauntletJsonOutput() (*GauntletResponse, error) {
	var payload = &GauntletResponse{}
	gauntletOutput, err := os.ReadFile(sg.dir + "report.json")
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(gauntletOutput, &payload)
	if err != nil {
		return payload, err
	}
	return payload, nil
}

// SetupNetwork Sets up a new network and sets the NODE_URL for Devnet / Starknet RPC
func (sg *StarknetGauntlet) SetupNetwork(addr string) error {
	sg.G.AddNetworkConfigVar("NODE_URL", addr)
	err := sg.G.WriteNetworkConfigMap(sg.dir + "packages-ts/starknet-gauntlet-cli/networks/")
	if err != nil {
		return err
	}

	return nil
}

func (sg *StarknetGauntlet) InstallDependencies() error {
	sg.G.Command = "yarn"
	_, err := sg.G.ExecCommand([]string{"install"}, *sg.options)
	if err != nil {
		return err
	}
	sg.G.Command = "gauntlet"
	return nil
}

func (sg *StarknetGauntlet) DeployAccountContract(salt int64, pubKey string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"account:deploy", fmt.Sprintf("--salt=%d", salt), fmt.Sprintf("--publicKey=%s", pubKey)}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) DeployLinkTokenContract() (string, error) {
	_, err := sg.G.ExecCommand([]string{"token:deploy", "--link"}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) MintLinkToken(token, to, amount string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"token:mint", fmt.Sprintf("--account=%s", to), fmt.Sprintf("--amount=%s", amount), token}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) TransferToken(token, to, amount string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"token:transfer", fmt.Sprintf("--recipient=%s", to), fmt.Sprintf("--amount=%s", amount), token}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) DeployOCR2ControllerContract(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, linkTokenAddress string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"ocr2:deploy", fmt.Sprintf("--minSubmissionValue=%d", minSubmissionValue), fmt.Sprintf("--maxSubmissionValue=%d", maxSubmissionValue), fmt.Sprintf("--decimals=%d", decimals), fmt.Sprintf("--name=%s", name), fmt.Sprintf("--link=%s", linkTokenAddress)}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) DeployAccessControllerContract() (string, error) {
	_, err := sg.G.ExecCommand([]string{"access_controller:deploy"}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) DeployOCR2ProxyContract(aggregator string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"proxy:deploy", fmt.Sprintf("--address=%s", aggregator)}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) SetOCRBilling(observationPaymentGjuels int64, transmissionPaymentGjuels int64, ocrAddress string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"ocr2:set_billing", fmt.Sprintf("--observationPaymentGjuels=%d", observationPaymentGjuels), fmt.Sprintf("--transmissionPaymentGjuels=%d", transmissionPaymentGjuels), ocrAddress}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) SetConfigDetails(cfg string, ocrAddress string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"ocr2:set_config", "--input=" + cfg, ocrAddress}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}

func (sg *StarknetGauntlet) AddAccess(aggregator, address string) (string, error) {
	_, err := sg.G.ExecCommand([]string{"ocr2:add_access", fmt.Sprintf("--address=%s", address), aggregator}, *sg.options)
	if err != nil {
		return "", err
	}
	sg.gr, err = sg.FetchGauntletJsonOutput()
	if err != nil {
		return "", err
	}
	return sg.gr.Responses[0].Contract, nil
}
