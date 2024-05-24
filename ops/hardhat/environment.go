package hardhat

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"

	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
)

type Chart struct {
	HelmProps *HelmProps
	Props     *Props
}
type Props struct {
	NetworkName string   `envconfig:"network_name"`
	Simulated   bool     `envconfig:"network_simulated"`
	HttpURLs    []string `envconfig:"http_url"` //nolint:revive
	WsURLs      []string `envconfig:"ws_url"`
	Values      map[string]any
}

type HelmProps struct {
	Name    string
	Path    string
	Values  *map[string]any
	Version string
}

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetVersion() string {
	return m.HelmProps.Version
}

func (m Chart) GetProps() any {
	return m.Props
}

func (m Chart) GetName() string {
	return m.HelmProps.Name
}

func (m Chart) GetPath() string {
	return m.HelmProps.Path
}

func (m Chart) GetValues() *map[string]any {
	return m.HelmProps.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	devnetLocalHTTP, err := e.Fwd.FindPort("hardhat:0", "hardhat", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	devnetInternalHTTP, err := e.Fwd.FindPort("hardhat:0", "hardhat", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetLocalHTTP)
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetInternalHTTP)
	log.Info().Str("Name", "Devnet").Str("URLs", devnetLocalHTTP).Msg("Devnet network")
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "hardhat",
		Values: map[string]any{
			"replicas": "1",
			"starknet-dev": map[string]any{
				"image": map[string]any{
					"image":   "ethereumoptimism/hardhat-node",
					"version": "nightly",
				},
				"resources": map[string]any{
					"requests": map[string]any{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
					"limits": map[string]any{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
				},
			},
		},
	}
}

func New(props *Props) environment.ConnectedChart {
	if props == nil {
		props = defaultProps()
	}
	return Chart{
		HelmProps: &HelmProps{
			Name:    "hardhat",
			Path:    fmt.Sprintf("%s/charts/hardhat", utils.OpsRoot),
			Values:  &props.Values,
			Version: "",
		},
		Props: props,
	}
}
