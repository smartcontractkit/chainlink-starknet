package testenv

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tc "github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"

	"github.com/smartcontractkit/chainlink-testing-framework/docker/test_env"
	"github.com/smartcontractkit/chainlink-testing-framework/logging"
	"github.com/smartcontractkit/chainlink-testing-framework/utils/testcontext"
)

const (
	StarkHTTPPort = "5050"
)

type Starknet struct {
	test_env.EnvComponent
	ExternalHTTPURL string
	InternalHTTPURL string
	t               *testing.T
	l               zerolog.Logger
	Image           string
}

func NewStarknet(networks []string, image string, opts ...test_env.EnvComponentOption) *Starknet {
	ms := &Starknet{
		Image: image,
		EnvComponent: test_env.EnvComponent{
			ContainerName: "starknet",
			Networks:      networks,
		},

		l: log.Logger,
	}
	for _, opt := range opts {
		opt(&ms.EnvComponent)
	}
	return ms
}

func (s *Starknet) WithTestLogger(t *testing.T) *Starknet {
	s.l = logging.GetTestLogger(t)
	s.t = t
	return s
}

func (s *Starknet) StartContainer() error {
	l := tc.Logger
	if s.t != nil {
		l = logging.CustomT{
			T: s.t,
			L: s.l,
		}
	}
	cReq, err := s.getContainerRequest()
	if err != nil {
		return err
	}
	c, err := tc.GenericContainer(testcontext.Get(s.t), tc.GenericContainerRequest{
		ContainerRequest: *cReq,
		Reuse:            true,
		Started:          true,
		Logger:           l,
	})
	if err != nil {
		return fmt.Errorf("cannot start Starknet container: %w", err)
	}
	s.Container = c
	host, err := test_env.GetHost(testcontext.Get(s.t), c)
	if err != nil {
		return err
	}
	httpPort, err := c.MappedPort(testcontext.Get(s.t), test_env.NatPort(StarkHTTPPort))
	if err != nil {
		return err
	}

	s.ExternalHTTPURL = fmt.Sprintf("http://%s:%s", host, httpPort.Port())
	s.InternalHTTPURL = fmt.Sprintf("http://%s:%s", s.ContainerName, StarkHTTPPort)

	s.l.Info().
		Any("ExternalHTTPURL", s.ExternalHTTPURL).
		Any("InternalHTTPURL", s.InternalHTTPURL).
		Str("containerName", s.ContainerName).
		Msgf("Started Starknet container")

	return nil
}

func (s *Starknet) getContainerRequest() (*tc.ContainerRequest, error) {
	return &tc.ContainerRequest{
		Name:         s.ContainerName,
		Image:        s.Image,
		ExposedPorts: []string{test_env.NatPortFormat(StarkHTTPPort)},
		Networks:     s.Networks,
		WaitingFor: tcwait.ForLog("Starknet Devnet listening").
			WithStartupTimeout(30 * time.Second).
			WithPollInterval(100 * time.Millisecond),
		Entrypoint: []string{"sh", "-c", "tini -- starknet-devnet --host 0.0.0.0 --port 5050 --seed 0 --account-class cairo1 --gas-price 1 --data-gas-price 1"},
	}, nil
}
