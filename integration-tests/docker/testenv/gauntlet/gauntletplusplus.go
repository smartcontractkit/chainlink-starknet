package testenv

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	tc "github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"

	"github.com/smartcontractkit/chainlink-testing-framework/lib/docker/test_env"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/logging"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/testcontext"
)

const (
	GauntletPlusPlusPort = "4444"
)

type GauntletPlusPlus struct {
	test_env.EnvComponent
	ExternalHTTPURL string
	InternalHTTPURL string
	t               *testing.T
	l               zerolog.Logger
	Image           string
}

func NewGauntletPlusPlus(networks []string, image string, opts ...test_env.EnvComponentOption) *GauntletPlusPlus {
	ms := &GauntletPlusPlus{
		Image: image,
		EnvComponent: test_env.EnvComponent{
			ContainerName: "gauntlet-plus-plus",
			Networks:      networks,
		},

		l: log.Logger,
	}

	for _, opt := range opts {
		opt(&ms.EnvComponent)
	}
	return ms
}

func (g *GauntletPlusPlus) WithTestLogger(t *testing.T) *GauntletPlusPlus {
	g.l = logging.GetTestLogger(t)
	g.t = t
	return g
}

func (g *GauntletPlusPlus) StartContainer() (string, error) {
	l := tc.Logger
	if g.t != nil {
		l = logging.CustomT{
			T: g.t,
			L: g.l,
		}
	}
	cReq, err := g.getContainerRequest()
	if err != nil {
		return "", err
	}
	c, err := tc.GenericContainer(testcontext.Get(g.t), tc.GenericContainerRequest{
		ContainerRequest: *cReq,
		Reuse:            true,
		Started:          true,
		Logger:           l,
	})
	if err != nil {
		return "", fmt.Errorf("cannot start GauntletPlusPlus container: %w", err)
	}

	g.Container = c
	host, err := test_env.GetHost(testcontext.Get(g.t), c)
	if err != nil {
		return "", err
	}

	httpPort, err := c.MappedPort(testcontext.Get(g.t), test_env.NatPort(GauntletPlusPlusPort))
	if err != nil {
		return "", err
	}

	g.ExternalHTTPURL = fmt.Sprintf("http://%s:%s", host, httpPort.Port())
	g.InternalHTTPURL = fmt.Sprintf("http://%s:%s", g.ContainerName, GauntletPlusPlusPort)

	g.l.Info().
		Any("ExternalHTTPURL", g.ExternalHTTPURL).
		Any("InternalHTTPURL", g.InternalHTTPURL).
		Str("containerName", g.ContainerName).
		Msgf("Started Gauntlet Plus Plus container")

	return g.ExternalHTTPURL, nil
}

func (g *GauntletPlusPlus) getContainerRequest() (*tc.ContainerRequest, error) {
	return &tc.ContainerRequest{
		Name:         g.ContainerName,
		Image:        g.Image,
		ExposedPorts: []string{test_env.NatPortFormat(GauntletPlusPlusPort)},
		Networks:     g.Networks,
		WaitingFor: tcwait.ForLog("Server listening at ").
			WithStartupTimeout(30 * time.Second).
			WithPollInterval(100 * time.Millisecond),
	}, nil
}
