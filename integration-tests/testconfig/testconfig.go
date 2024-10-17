package testconfig

import (
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/barkimedes/go-deepcopy"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/smartcontractkit/chainlink/integration-tests/types/config/node"

	common_cfg "github.com/smartcontractkit/chainlink-common/pkg/config"

	ctf_config "github.com/smartcontractkit/chainlink-testing-framework/lib/config"
	k8s_config "github.com/smartcontractkit/chainlink-testing-framework/lib/k8s/config"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/logging"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/osutil"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/ptr"
	"github.com/smartcontractkit/chainlink-testing-framework/seth"

	ocr2_config "github.com/smartcontractkit/chainlink-starknet/integration-tests/testconfig/ocr2"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/config"
)

type TestConfig struct {
	ChainlinkImage        *ctf_config.ChainlinkImageConfig `toml:"ChainlinkImage"`
	Logging               *ctf_config.LoggingConfig        `toml:"Logging"`
	ChainlinkUpgradeImage *ctf_config.ChainlinkImageConfig `toml:"ChainlinkUpgradeImage"`
	Network               *ctf_config.NetworkConfig        `toml:"Network"`
	Common                *Common                          `toml:"Common"`
	OCR2                  *ocr2_config.Config              `toml:"OCR2"`
	ConfigurationName     string                           `toml:"-"`

	// getter funcs for passing parameters
	GetChainID, GetFeederURL, GetRPCL2Internal, GetRPCL2InternalAPIKey func() string
}

func (c *TestConfig) GetLoggingConfig() *ctf_config.LoggingConfig {
	return c.Logging
}

func (c *TestConfig) GetPrivateEthereumNetworkConfig() *ctf_config.EthereumNetworkConfig {
	return &ctf_config.EthereumNetworkConfig{}
}

func (c *TestConfig) GetSethConfig() *seth.Config {
	return nil
}

func (c *TestConfig) GetPyroscopeConfig() *ctf_config.PyroscopeConfig {
	return &ctf_config.PyroscopeConfig{}
}

var embeddedConfigs embed.FS
var areConfigsEmbedded bool

func init() {
	embeddedConfigs = embeddedConfigsFs
}

// Saves Test Config to a local file
func (c *TestConfig) Save() (string, error) {
	filePath := fmt.Sprintf("test_config-%s.toml", uuid.New())

	content, err := toml.Marshal(*c)
	if err != nil {
		return "", fmt.Errorf("error marshaling test config: %w", err)
	}

	err = os.WriteFile(filePath, content, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing test config: %w", err)
	}

	return filePath, nil
}

// MustCopy Returns a deep copy of the Test Config or panics on error
func (c TestConfig) MustCopy() any {
	return deepcopy.MustAnything(c).(TestConfig)
}

// MustCopy Returns a deep copy of struct passed to it and returns a typed copy (or panics on error)
func MustCopy[T any](c T) T {
	return deepcopy.MustAnything(c).(T)
}

func (c TestConfig) GetNetworkConfig() *ctf_config.NetworkConfig {
	return c.Network
}

func (c TestConfig) GetChainlinkImageConfig() *ctf_config.ChainlinkImageConfig {
	return c.ChainlinkImage
}

func (c TestConfig) GetCommonConfig() *Common {
	return c.Common
}

func (c *TestConfig) GetNodeConfig() *ctf_config.NodeConfig {
	cfgTOML, err := c.GetNodeConfigTOML()
	if err != nil {
		log.Fatalf("failed to parse TOML config: %s", err)
		return nil
	}

	return &ctf_config.NodeConfig{
		BaseConfigTOML: cfgTOML,
	}
}

func (c TestConfig) GetNodeConfigTOML() (string, error) {
	var chainID, feederURL, RPCL2Internal, RPCL2InternalAPIKey string
	if c.GetChainID != nil {
		chainID = c.GetChainID()
	}
	if c.GetFeederURL != nil {
		feederURL = c.GetFeederURL()
	}
	if c.GetRPCL2Internal != nil {
		RPCL2Internal = c.GetRPCL2Internal()
	}
	if c.GetRPCL2InternalAPIKey != nil {
		RPCL2InternalAPIKey = c.GetRPCL2InternalAPIKey()
	}

	starkConfig := config.TOMLConfig{
		Enabled:   ptr.Ptr(true),
		ChainID:   ptr.Ptr(chainID),
		FeederURL: common_cfg.MustParseURL(feederURL),
		Nodes: []*config.Node{
			{
				Name:   ptr.Ptr("primary"),
				URL:    common_cfg.MustParseURL(RPCL2Internal),
				APIKey: ptr.Ptr(RPCL2InternalAPIKey),
			},
		},
	}
	baseConfig := node.NewBaseConfig()
	baseConfig.Starknet = config.TOMLConfigs{
		&starkConfig,
	}
	baseConfig.OCR2.Enabled = ptr.Ptr(true)
	baseConfig.P2P.V2.Enabled = ptr.Ptr(true)
	fiveSecondDuration := common_cfg.MustNewDuration(5 * time.Second)

	baseConfig.P2P.V2.DeltaDial = fiveSecondDuration
	baseConfig.P2P.V2.DeltaReconcile = fiveSecondDuration
	baseConfig.P2P.V2.ListenAddresses = &[]string{"0.0.0.0:6690"}

	return baseConfig.TOMLString()
}

func (c TestConfig) GetChainlinkUpgradeImageConfig() *ctf_config.ChainlinkImageConfig {
	return c.ChainlinkUpgradeImage
}

func (c TestConfig) GetConfigurationName() string {
	return c.ConfigurationName
}

func (c *TestConfig) AsBase64() (string, error) {
	content, err := toml.Marshal(*c)
	if err != nil {
		return "", fmt.Errorf("error marshaling test config: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

type Common struct {
	Network   *string `toml:"network"`
	InsideK8s *bool   `toml:"inside_k8"`
	User      *string `toml:"user"`
	// if rpc requires api key to be passed as an HTTP header
	L2RPCApiKey        *string `toml:"l2_rpc_url_api_key"`
	L2RPCUrl           *string `toml:"l2_rpc_url"`
	PrivateKey         *string `toml:"private_key"`
	Account            *string `toml:"account"`
	Stateful           *bool   `toml:"stateful_db"`
	InternalDockerRepo *string `toml:"internal_docker_repo"`
	DevnetImage        *string `toml:"devnet_image"`
	GauntletPlusPlusImage *string `toml:"gauntlet_plus_plus_image"`
	PostgresVersion    *string `toml:"postgres_version"`
	GauntletPlusPlusPort *string `toml:"gauntlet_plus_plus_port"`
}

func (c *Common) Validate() error {
	if c.Network == nil {
		return fmt.Errorf("network must be set")
	}

	switch *c.Network {
	case "localnet":
		if c.DevnetImage == nil {
			return fmt.Errorf("devnet_image must be set")
		}
		if c.GauntletPlusPlusImage == nil {
			return fmt.Errorf("gauntlet_plus_plus_image must be set")
		}
	case "testnet":
		if c.PrivateKey == nil {
			return fmt.Errorf("private_key must be set")
		}
		if c.L2RPCUrl == nil {
			return fmt.Errorf("l2_rpc_url must be set")
		}

		if c.Account == nil {
			return fmt.Errorf("account must be set")
		}
	default:
		return fmt.Errorf("network must be either 'localnet' or 'testnet'")
	}

	if c.InsideK8s == nil {
		return fmt.Errorf("inside_k8 must be set")
	}

	if c.InternalDockerRepo == nil {
		return fmt.Errorf("internal_docker_repo must be set")
	}

	if c.User == nil {
		return fmt.Errorf("user must be set")
	}

	if c.Stateful == nil {
		return fmt.Errorf("stateful_db state for db must be set")
	}

	if c.PostgresVersion == nil {
		return fmt.Errorf("postgres_version must be set")
	}

	return nil
}

type Product string

const (
	OCR2 Product = "ocr2"
)

const TestTypeEnvVarName = "TEST_TYPE"

const (
	Base64OverrideEnvVarName = k8s_config.EnvBase64ConfigOverride
	NoKey                    = "NO_KEY"
)

func GetConfig(configurationName string, product Product) (TestConfig, error) {
	logger := logging.GetTestLogger(nil)

	configurationName = strings.ReplaceAll(configurationName, "/", "_")
	configurationName = strings.ReplaceAll(configurationName, " ", "_")
	configurationName = cases.Title(language.English, cases.NoLower).String(configurationName)
	fileNames := []string{
		"default.toml",
		fmt.Sprintf("%s.toml", product),
		"overrides.toml",
	}

	testConfig := TestConfig{}
	testConfig.ConfigurationName = configurationName
	logger.Debug().Msgf("Will apply configuration named '%s' if it is found in any of the configs", configurationName)

	var handleSpecialOverrides = func(logger zerolog.Logger, filename, configurationName string, target *TestConfig, content []byte, product Product) error {
		switch product {
		default:
			err := ctf_config.BytesToAnyTomlStruct(logger, filename, configurationName, target, content)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", filename, err)
			}

			return nil
		}
	}

	// read embedded configs is build tag "embed" is set
	// this makes our life much easier when using a binary
	if areConfigsEmbedded {
		logger.Info().Msg("Reading embedded configs")
		embeddedFiles := []string{"default.toml", fmt.Sprintf("%s/%s.toml", product, product)}
		for _, fileName := range embeddedFiles {
			file, err := embeddedConfigs.ReadFile(fileName)
			if err != nil && errors.Is(err, os.ErrNotExist) {
				logger.Debug().Msgf("Embedded config file %s not found. Continuing", fileName)
				continue
			} else if err != nil {
				return TestConfig{}, fmt.Errorf("error reading embedded config: %w", err)
			}

			err = handleSpecialOverrides(logger, fileName, "", &testConfig, file, product) // use empty configurationName to read default config
			if err != nil {
				return TestConfig{}, fmt.Errorf("error unmarshalling embedded config: %w", err)
			}
		}
	}

	logger.Info().Msg("Reading configs from file system")
	for _, fileName := range fileNames {
		logger.Debug().Msgf("Looking for config file %s", fileName)
		filePath, err := osutil.FindFile(fileName, osutil.DEFAULT_STOP_FILE_NAME, 3)

		if err != nil && errors.Is(err, os.ErrNotExist) {
			logger.Debug().Msgf("Config file %s not found", fileName)
			continue
		} else if err != nil {
			return TestConfig{}, fmt.Errorf("error looking for file %s: %w", filePath, err)
		}
		logger.Debug().Str("location", filePath).Msgf("Found config file %s", fileName)

		content, err := readFile(filePath)
		if err != nil {
			return TestConfig{}, fmt.Errorf("error reading file %s: %w", filePath, err)
		}

		err = handleSpecialOverrides(logger, fileName, "", &testConfig, content, product) // use empty configurationName to read default config
		if err != nil {
			return TestConfig{}, fmt.Errorf("error reading file %s: %w", filePath, err)
		}
	}

	logger.Info().Msg("Reading configs from Base64 override env var")
	configEncoded, isSet := os.LookupEnv(Base64OverrideEnvVarName)
	if isSet && configEncoded != "" {
		logger.Debug().Msgf("Found base64 config override environment variable '%s' found", Base64OverrideEnvVarName)
		decoded, err := base64.StdEncoding.DecodeString(configEncoded)
		if err != nil {
			return TestConfig{}, err
		}

		err = handleSpecialOverrides(logger, Base64OverrideEnvVarName, "", &testConfig, decoded, product) // use empty configurationName to read default config
		if err != nil {
			return TestConfig{}, fmt.Errorf("error unmarshaling base64 config: %w", err)
		}
	} else {
		logger.Debug().Msg("Base64 config override from environment variable not found")
	}

	// it neede some custom logic, so we do it separately
	err := testConfig.readNetworkConfiguration()
	if err != nil {
		return TestConfig{}, fmt.Errorf("error reading network config: %w", err)
	}
	testConfig.ReadEnvVars()

	logger.Debug().Msg("Validating test config")
	err = testConfig.Validate()
	if err != nil {
		return TestConfig{}, fmt.Errorf("error validating test config: %w", err)
	}

	if testConfig.Common == nil {
		testConfig.Common = &Common{}
	}

	logger.Debug().Msg("Correct test config constructed successfully")
	return testConfig, nil
}

func (c *TestConfig) readNetworkConfiguration() error {
	// currently we need to read that kind of secrets only for network configuration
	if c.Network == nil {
		c.Network = &ctf_config.NetworkConfig{}
	}
	c.Network.UpperCaseNetworkNames()
	return nil
}

func (c *TestConfig) ReadEnvVars() {
	image := ctf_config.MustReadEnvVar_String("CHAINLINK_IMAGE")
	c.ChainlinkImage.Image = &image
	version := ctf_config.MustReadEnvVar_String("CHAINLINK_VERSION")
	c.ChainlinkImage.Version = &version
}

func (c *TestConfig) Validate() error {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("Panic during test config validation: '%v'. Most probably due to presence of partial product config", r))
		}
	}()
	if c.ChainlinkImage == nil {
		return fmt.Errorf("chainlink image config must be set")
	}
	if err := c.ChainlinkImage.Validate(); err != nil {
		return fmt.Errorf("chainlink image config validation failed: %w", err)
	}
	if c.ChainlinkUpgradeImage != nil {
		if err := c.ChainlinkUpgradeImage.Validate(); err != nil {
			return fmt.Errorf("chainlink upgrade image config validation failed: %w", err)
		}
	}
	if err := c.Network.Validate(); err != nil {
		return fmt.Errorf("network config validation failed: %w", err)
	}

	if c.Common == nil {
		return fmt.Errorf("common config must be set")
	}

	if err := c.Common.Validate(); err != nil {
		return fmt.Errorf("Common config validation failed: %w", err)
	}

	if c.OCR2 == nil {
		return fmt.Errorf("OCR2 config must be set")
	}

	if err := c.OCR2.Validate(); err != nil {
		return fmt.Errorf("OCR2 config validation failed: %w", err)
	}
	return nil
}

func readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return content, nil
}
