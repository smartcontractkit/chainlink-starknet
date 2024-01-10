package devnet

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/smartcontractkit/chainlink-testing-framework/k8s/client"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/config"
	"github.com/smartcontractkit/chainlink-testing-framework/k8s/environment"

	"github.com/smartcontractkit/chainlink-starknet/ops/utils"
)

const NetworkName = "starknet-dev"

type Chart struct {
	HelmProps *HelmProps
	Props     *Props
}
type Props struct {
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

func (m Chart) GetVersion() string {
	return m.HelmProps.Version
}

func (m Chart) ExportData(e *environment.Environment) error {
	devnetLocalHttp, err := e.Fwd.FindPort("starknet-dev:0", "starknetdev", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	devnetInternalHttp, err := e.Fwd.FindPort("starknet-dev:0", "starknetdev", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	e.URLs[NetworkName] = append(e.URLs[NetworkName], devnetLocalHttp)
	e.URLs[NetworkName] = append(e.URLs[NetworkName], devnetInternalHttp)
	log.Info().Str("Name", "Devnet").Str("URLs", devnetLocalHttp).Msg("Devnet network")
	return nil
}

func defaultProps() map[string]any {
	return map[string]any{
		"replicas": "1",
		"starknet-dev": map[string]any{
			"image": map[string]any{
				"image":   "shardlabs/starknet-devnet-rs",
				"version": "5d2536a99852b1a61bbbfdcaa6755cb4275bffddm",
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
			"seed":      "123",
			"real_node": "false",
		},
	}

}

func New(props *Props) environment.ConnectedChart {
	dp := defaultProps()
	if props != nil {
		config.MustMerge(&dp, props)
	}

	return Chart{
		HelmProps: &HelmProps{
			Name:   NetworkName,
			Path:   fmt.Sprintf("%s/charts/devnet", utils.OpsRoot),
			Values: &dp,
		},
		Props: props,
	}
}
