package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/utils"
)

func (m *OCRv2TestState) fundNodesWithGPP() ([]string, error) {
	l := utils.GetTestLogger(m.TestConfig.T)
	var nAccounts []string
	err := m.Clients.GauntletPPClient.DeclareOzAccount()
	if err != nil {
		return nil, err
	}
	for _, key := range m.GetNodeKeys() {
		if key.TXKey.Data.Attributes.StarkKey == "" {
			return nil, errors.New("stark key can't be empty")
		}
		nAccount, err := m.Clients.GauntletPPClient.DeployOzAccount(key.TXKey.Data.Attributes.StarkKey)
		if err != nil {
			return nil, err
		}
		nAccounts = append(nAccounts, nAccount)
	}

	if *m.Common.TestConfig.Common.Network == "testnet" {
		for _, key := range nAccounts {
			// We are not deploying in parallel here due to testnet limitations (429 too many requests)
			l.Debug().Msg(fmt.Sprintf("Funding node with address: %s", key))
			err := m.Clients.GauntletPPClient.TransferToken(m.Common.ChainDetails.StarkTokenAddress, key, "10000000000000000000")
			if err != nil {
				return nil, err
			}
		}
	} else {
		// The starknet provided mint method does not work so we send a req directly
		for _, key := range nAccounts {
			res, err := m.TestConfig.Resty.R().SetBody(map[string]any{
				"address": key,
				"amount":  900000000000000000,
			}).Post("/mint")
			if err != nil {
				return nil, err
			}
			l.Info().Msg(fmt.Sprintf("Funding account (WEI): %s", string(res.Body())))
			res, err = m.TestConfig.Resty.R().SetBody(map[string]any{
				"address": key,
				"amount":  900000000000000000,
				"unit":    m.Common.ChainDetails.TokenName,
			}).Post("/mint")
			if err != nil {
				return nil, err
			}
			l.Info().Msg(fmt.Sprintf("Funding account (FRI): %s", string(res.Body())))
		}
	}

	return nAccounts, nil
}

func (m *OCRv2TestState) deployAccessControllerWithGpp() error {
	err := m.Clients.GauntletPPClient.DeclareAccessControllerContract()
	if err != nil {
		return err
	}

	m.Contracts.AccessControllerAddr, err = m.Clients.GauntletPPClient.DeployAccessControllerContract(m.Account.Account)
	if err != nil {
		return err
	}
	err = os.Setenv("BILLING_ACCESS_CONTROLLER", m.Contracts.AccessControllerAddr)
	if err != nil {
		return err
	}
	return nil
}

func (m *OCRv2TestState) deployLinkTokenWithGpp() error {
	var err error
	m.Contracts.LinkTokenAddr, err = m.Clients.GauntletPPClient.DeployLinkTokenContract(m.Account.Account)

	if err != nil {
		return err
	}

	err = os.Setenv("LINK", m.Contracts.LinkTokenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (m *OCRv2TestState) setConfigDetailsWithGpp(ocrAddress string) error {
	cfg, err := m.LoadOCR2Config()
	if err != nil {
		return err
	}
	var parsedConfig []byte
	parsedConfig, err = json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = m.Clients.GauntletPPClient.SetConfigDetails(string(parsedConfig), ocrAddress)
	return err
}

func (m *OCRv2TestState) setConfigDetails(ocrAddress string) error {
	cfg, err := m.LoadOCR2Config()
	if err != nil {
		return err
	}
	var parsedConfig []byte
	parsedConfig, err = json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = m.Clients.GauntletClient.SetConfigDetails(string(parsedConfig), ocrAddress)
	return err
}

func (m *OCRv2TestState) DeployGauntletPP(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, observationPaymentGjuels int64, transmissionPaymentGjuels int64) error {
	err := m.Clients.GauntletClient.InstallDependencies()
	if err != nil {
		return err
	}

	m.Clients.ChainlinkClient.AccountAddresses, err = m.fundNodesWithGPP()
	if err != nil {
		return err
	}

	err = m.deployLinkTokenWithGpp()
	if err != nil {
		return err
	}

	err = m.deployAccessControllerWithGpp()
	if err != nil {
		return err
	}

	m.Contracts.OCRAddr, err = m.Clients.GauntletPPClient.DeployOCR2ControllerContract(minSubmissionValue, maxSubmissionValue, decimals, name,
		m.Contracts.LinkTokenAddr, m.Account.Account, m.Contracts.AccessControllerAddr)
	if err != nil {
		return err
	}

	m.Contracts.ProxyAddr, err = m.Clients.GauntletPPClient.DeployOCR2ControllerProxyContract(m.Account.Account, m.Contracts.OCRAddr)
	if err != nil {
		return err
	}

	err = m.Clients.GauntletPPClient.AddAccess(m.Contracts.OCRAddr, m.Contracts.ProxyAddr)
	if err != nil {
		return err
	}

	// Gauntlet PP does not have a mint op. We will use legacy gauntlet until we implement one
	_, err = m.Clients.GauntletClient.MintLinkToken(m.Contracts.LinkTokenAddr, m.Contracts.OCRAddr, "100000000000000000000")
	if err != nil {
		return err
	}

	_, err = m.Clients.GauntletPPClient.SetOCRBilling(observationPaymentGjuels, transmissionPaymentGjuels, m.Contracts.OCRAddr)
	if err != nil {
		return err
	}

	// Gauntlet PP set-config op has different functionality than legacy
	// TODO: Investigate how to set config using GPP
	err = m.setConfigDetails(m.Contracts.OCRAddr)
	return err
}
