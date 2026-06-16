package testenv

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcwait "github.com/testcontainers/testcontainers-go/wait"

	"github.com/smartcontractkit/chainlink-testing-framework/lib/docker/test_env"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/logging"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/testcontext"
)

const GauntletPlusPlusPort = "4444"

type GauntletPlusPlus struct {
	test_env.EnvComponent
	ExternalHTTPURL string
	InternalHTTPURL string
	t               *testing.T
	l               zerolog.Logger
	Version         string
	installDir      string
}

func NewGauntletPlusPlus(networks []string, version string, opts ...test_env.EnvComponentOption) *GauntletPlusPlus {
	ms := &GauntletPlusPlus{
		Version: version,
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
	installDir, err := gauntletPlusPlusInstallDir()
	if err != nil {
		return "", fmt.Errorf("prepare gauntlet++ release v%s: %w", g.Version, err)
	}
	g.installDir = installDir

	l := tclog.Default()
	if g.t != nil {
		l = logging.CustomT{
			T: g.t,
			L: g.l,
		}
	}

	cReq, err := g.getContainerRequest(installDir)
	if err != nil {
		return "", err
	}

	c, err := testcontainers.GenericContainer(testcontext.Get(g.t), testcontainers.GenericContainerRequest{
		ContainerRequest: *cReq,
		// Fresh G++ install each run; reuse caused stale plugin state across smoke runs.
		Reuse:   false,
		Started: true,
		Logger:  l,
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
		Str("version", g.Version).
		Str("installDir", installDir).
		Any("ExternalHTTPURL", g.ExternalHTTPURL).
		Any("InternalHTTPURL", g.InternalHTTPURL).
		Str("containerName", g.ContainerName).
		Msg("Started Gauntlet Plus Plus from nops tarball")

	return g.ExternalHTTPURL, nil
}

func gauntletPlusPlusInstallDir() (string, error) {
	installDir := os.Getenv("GAUNTLET_PLUS_PLUS_DIR")
	if installDir == "" {
		return "", fmt.Errorf(
			"GAUNTLET_PLUS_PLUS_DIR is not set; run integration-tests/scripts/download-gauntlet-plus-plus.sh before tests",
		)
	}

	gauntletBin := filepath.Join(installDir, "bin", "gauntlet")
	if _, err := os.Stat(gauntletBin); err != nil {
		return "", fmt.Errorf("GAUNTLET_PLUS_PLUS_DIR=%s missing bin/gauntlet: %w", installDir, err)
	}

	return installDir, nil
}

func (g *GauntletPlusPlus) getContainerRequest(installDir string) (*testcontainers.ContainerRequest, error) {
	mount := fmt.Sprintf("%s:/gauntlet:rw", installDir)

	return &testcontainers.ContainerRequest{
		Name:          g.ContainerName,
		Image:         "node:18-bookworm",
		ImagePlatform: "linux/amd64",
		ExposedPorts:  []string{test_env.NatPortFormat(GauntletPlusPlusPort)},
		Networks:      g.Networks,
		WorkingDir:    "/gauntlet",
		Env: map[string]string{
			"GAUNTLET_DATA_DIR":   "/gauntlet/data",
			"GAUNTLET_CONFIG_DIR": "/gauntlet/config",
			"GAUNTLET_CACHE_DIR":  "/gauntlet/cache",
		},
		Cmd: []string{
			"/gauntlet/bin/gauntlet",
			"serve",
			"-h", "0.0.0.0",
			"-p", GauntletPlusPlusPort,
		},
		WaitingFor: tcwait.ForLog("Server listening at ").
			WithStartupTimeout(5 * time.Minute).
			WithPollInterval(100 * time.Millisecond),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = append(hc.Binds, mount)
		},
	}, nil
}
