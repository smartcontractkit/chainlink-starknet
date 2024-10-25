package common

import (
	"encoding/json"
	"os"
)

func (m *OCRv2TestState) deployAccessControllerWithGpp() error {
	var err error
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

func (m *OCRv2TestState) declareLinkToken() error {
	err := m.Clients.GauntletPPClient.DeclareLinkTokenContract()
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
