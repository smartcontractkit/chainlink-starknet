package devnet

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/client"
	"github.com/smartcontractkit/chainlink-env/environment"
)

type Chart struct {
	HelmProps *HelmProps
	Props     *Props
}
type Props struct {
	NetworkName string   `envconfig:"network_name"`
	Simulated   bool     `envconfig:"network_simulated"`
	HttpURLs    []string `envconfig:"http_url"`
	WsURLs      []string `envconfig:"ws_url"`
	Values      map[string]interface{}
}

type HelmProps struct {
	Name    string
	Path    string
	Values  *map[string]interface{}
	Version string
}

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
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetLocalHttp)
	e.URLs[m.Props.NetworkName] = append(e.URLs[m.Props.NetworkName], devnetInternalHttp)
	log.Info().Str("Name", "Devnet").Str("URLs", devnetLocalHttp).Msg("Devnet network")
	return nil
}

func defaultProps() *Props {
	return &Props{
		NetworkName: "starknet-dev",
		Values: map[string]interface{}{
			"replicas": "1",
			"starknet-dev": map[string]interface{}{
				"image": map[string]interface{}{
					"image":   "shardlabs/starknet-devnet",
					"version": "v0.3.5",
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
				"seed":      "123",
				"real_node": "false",
			},
		},
	}
}

func New(helmVersion string, props *Props) environment.ConnectedChart {
	defaultPath := "../../ops/charts/devnet"
	_, InsideK8s := os.LookupEnv("ENV_INSIDE_K8S")
	if InsideK8s {
		defaultPath = "/repo/ops/charts/devnet"
	}
	if props == nil {
		props = defaultProps()
	}

	return Chart{
		HelmProps: &HelmProps{
			Name:    "starknet-dev",
			Path:    defaultPath,
			Values:  &props.Values,
			Version: helmVersion,
		},
		Props: props,
	}
}
