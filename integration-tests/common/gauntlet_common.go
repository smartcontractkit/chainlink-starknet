package common

import (
	"github.com/smartcontractkit/chainlink-starknet/ops/devnet"
	"os"
)

func (t *Test) FundNodes() error {
	var nAccounts []string
	for _, key := range t.GetNodeKeys() {
		if key.TXKey.Data.Attributes.StarkKey == "" {
			return err
		}
		nAccount, err := t.Sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
		if err != nil || nAccount != key.TXKey.Data.Attributes.AccountAddr {
			return err
		}
		nAccounts = append(nAccounts, nAccount)
	}
	err = devnet.FundAccounts(nAccounts)
	if err != nil {
		return err
	}

	return nil
}

func (t *Test) DeployLinkToken() (string, error) {
	linkTokenAddress, err := t.Sg.DeployLinkTokenContract()
	if err != nil {
		return "", err
	}
	err = os.Setenv("LINK", linkTokenAddress)
	if err != nil {
		return "", err
	}
	return linkTokenAddress, nil
}

func (t *Test) DeployAccessController() (string, error) {
	accessControllerAddress, err := t.Sg.DeployAccessControllerContract()
	if err != nil {
		return "", err
	}
	err = os.Setenv("BILLING_ACCESS_CONTROLLER", accessControllerAddress)
	if err != nil {
		return "", err
	}
	return accessControllerAddress, nil
}

//func (t *Test) DeployOCR2ControllerContract(minSubmissionValue int64, maxSubmissionValue int64) (string, error) {
//	ocrAddress, err := t.Sg.DeployOCR2ControllerContract(minSubmissionValue, maxSubmissionValue, decimals, "auto", linkTokenAddress)
//	if err != nil {
//		return "", err
//	}
//	err = os.Setenv("BILLING_ACCESS_CONTROLLER", ocrAddress)
//	if err != nil {
//		return "", err
//	}
//	return ocrAddress, nil
//}
