package common

import (
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
	var err error
	err = m.Clients.GauntletPPClient.DeclareLinkTokenContract()
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
