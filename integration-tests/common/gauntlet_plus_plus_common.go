package common

import (
	"os"
)

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
