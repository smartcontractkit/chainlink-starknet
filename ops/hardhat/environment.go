package hardhat

import (
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
	"github.com/smartcontractkit/chainlink-env/pkg/helm/ethereum"
)

type Chart ethereum.Chart
type Props ethereum.Props
type HelmProps ethereum.HelmProps

func (m Chart) IsDeploymentNeeded() bool {
	return true
}

func (m Chart) GetProps() interface{} {
	return m.Props
}

func (m Chart) GetName() string {
	return m.HelmProps.Name
}

func (m Chart) GetPath() string {
	return m.HelmProps.Path
}

func (m Chart) GetValues() *map[string]interface{} {
	return m.HelmProps.Values
}

func (m Chart) ExportData(e *environment.Environment) error {
	devnetLocalHttp, err := e.Fwd.FindPort("hardhat:0", "hardhat", "http").As(client.LocalConnection, client.HTTP)
	if err != nil {
		return err
	}
	devnetInternalHttp, err := e.Fwd.FindPort("hardhat:0", "hardhat", "http").As(client.RemoteConnection, client.HTTP)
	if err != nil {
		return err
	}
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetLocalHttp)
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetInternalHttp)
	log.Info().Str("Name", "Devnet").Str("URLs", devnetLocalHttp).Msg("Devnet network")
	return nil
}

func defaultProps() *ethereum.Props {
	return &ethereum.Props{
		NetworkName: "hardhat",
		Values: map[string]interface{}{
			"replicas": "1",
			"starknet-dev": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   "ethereumoptimism/hardhat-node",
					"version": "nightly",
				},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "1000m",
						"memory": "1024Mi",
					},
				},
			},
		},
	}
}

func New(props *ethereum.Props) environment.ConnectedChart {
	if props == nil {
		props = defaultProps()
	}
	return Chart{
		HelmProps: &ethereum.HelmProps{
			Name:   "hardhat",
			Path:   "../../ops/charts/hardhat",
			Values: &props.Values,
		},
		Props: props,
	}
}
