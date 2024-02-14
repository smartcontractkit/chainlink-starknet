package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/integration-tests/utils"
	"os"
)

var (
	ethAddressGoerli = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	nAccount         string
)

func (m *OCRv2TestState) fundNodes() ([]string, error) {
	l := utils.GetTestLogger(m.T)
	var nAccounts []string
	var err error
	for _, key := range m.GetNodeKeys() {
		if key.TXKey.Data.Attributes.StarkKey == "" {
			return nil, errors.New("stark key can't be empty")
		}
		nAccount, err = m.Sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
		if err != nil {
			return nil, err
		}
		nAccounts = append(nAccounts, nAccount)
	}

	if err != nil {
		return nil, err
	}

	if m.Common.Testnet {
		for _, key := range nAccounts {
			// We are not deploying in parallel here due to testnet limitations (429 too many requests)
			l.Debug().Msg(fmt.Sprintf("Funding node with address: %s", key))
			_, err = m.Sg.TransferToken(ethAddressGoerli, key, "100000000000000000") // Transferring 1 ETH to each node
			if err != nil {
				return nil, err
			}
		}

	} else {
		// The starknet provided mint method does not work so we send a req directly
		for _, key := range nAccounts {
			res, err := m.Resty.R().SetBody(map[string]any{
				"address": key,
				"amount":  900000000000000000,
			}).Post("/mint")
			m.L.Info().Msg(fmt.Sprintf("Funding account: %s", string(res.Body())))
			if err != nil {
				return nil, err
			}
		}
	}

	return nAccounts, nil
}

func (m *OCRv2TestState) deployLinkToken() error {
	var err error
	m.LinkTokenAddr, err = m.Sg.DeployLinkTokenContract()
	if err != nil {
		return err
	}
	err = os.Setenv("LINK", m.LinkTokenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (m *OCRv2TestState) deployAccessController() error {
	var err error
	m.AccessControllerAddr, err = m.Sg.DeployAccessControllerContract()
	if err != nil {
		return err
	}
	err = os.Setenv("BILLING_ACCESS_CONTROLLER", m.AccessControllerAddr)
	if err != nil {
		return err
	}
	return nil
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
	_, err = m.Sg.SetConfigDetails(string(parsedConfig), ocrAddress)
	return err
}

func (m *OCRv2TestState) DeployGauntlet(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, observationPaymentGjuels int64, transmissionPaymentGjuels int64) error {
	err := m.Sg.InstallDependencies()
	if err != nil {
		return err
	}

	m.AccountAddresses, err = m.fundNodes()
	if err != nil {
		return err
	}

	err = m.deployLinkToken()
	if err != nil {
		return err
	}

	err = m.deployAccessController()
	if err != nil {
		return err
	}

	m.OCRAddr, err = m.Sg.DeployOCR2ControllerContract(minSubmissionValue, maxSubmissionValue, decimals, name, m.LinkTokenAddr)
	if err != nil {
		return err
	}

	m.ProxyAddr, err = m.Sg.DeployOCR2ProxyContract(m.OCRAddr)
	if err != nil {
		return err
	}
	_, err = m.Sg.AddAccess(m.OCRAddr, m.ProxyAddr)
	if err != nil {
		return err
	}

	_, err = m.Sg.MintLinkToken(m.LinkTokenAddr, m.OCRAddr, "100000000000000000000")
	if err != nil {
		return err
	}
	_, err = m.Sg.SetOCRBilling(observationPaymentGjuels, transmissionPaymentGjuels, m.OCRAddr)
	if err != nil {
		return err
	}

	err = m.setConfigDetails(m.OCRAddr)
	return err
}
