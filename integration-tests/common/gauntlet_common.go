package common

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

var (
	ethAddressGoerli = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	nAccount         string
)

func (t *Test) fundNodes() error {
	var nAccounts []string
	for _, key := range t.GetNodeKeys() {
		if key.TXKey.Data.Attributes.StarkKey == "" {
			return err
		}
		nAccount, err = t.Sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
		if err != nil || nAccount != key.TXKey.Data.Attributes.AccountAddr {
			return err
		}
		nAccounts = append(nAccounts, nAccount)
	}
	if t.Testnet {
		for _, key := range nAccounts {
			log.Debug().Msg(fmt.Sprintf("Funding node with address: %s", key))
			_, err = t.Sg.TransferToken(ethAddressGoerli, key, "10000000000000000") // Transferring 0.1 ETH to each node
			if err != nil {
				return err
			}
		}
	} else {
		err = t.Devnet.FundAccounts(nAccounts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Test) deployLinkToken() error {
	t.LinkTokenAddr, err = t.Sg.DeployLinkTokenContract()
	if err != nil {
		return err
	}
	err = os.Setenv("LINK", t.LinkTokenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (t *Test) deployAccessController() error {
	t.AccessControllerAddr, err = t.Sg.DeployAccessControllerContract()
	if err != nil {
		return err
	}
	err = os.Setenv("BILLING_ACCESS_CONTROLLER", t.AccessControllerAddr)
	if err != nil {
		return err
	}
	return nil
}

func (t *Test) setConfigDetails(ocrAddress string) error {
	cfg, err := t.LoadOCR2Config()
	if err != nil {
		return err
	}
	parsedConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = t.Sg.SetConfigDetails(string(parsedConfig), ocrAddress)
	return nil
}

func (t *Test) DeployGauntlet(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, observationPaymentGjuels int64, transmissionPaymentGjuels int64) error {
	err = t.fundNodes()
	if err != nil {
		return err
	}

	err = t.deployLinkToken()
	if err != nil {
		return err
	}

	err = t.deployAccessController()
	if err != nil {
		return err
	}

	t.OCRAddr, err = t.Sg.DeployOCR2ControllerContract(minSubmissionValue, maxSubmissionValue, decimals, name, t.LinkTokenAddr)
	if err != nil {
		return err
	}

	t.ProxyAddr, err = t.Sg.DeployOCR2ProxyContract(t.OCRAddr)
	if err != nil {
		return err
	}

	_, err = t.Sg.MintLinkToken(t.LinkTokenAddr, t.OCRAddr, "100000000000000000000")
	if err != nil {
		return err
	}
	_, err = t.Sg.SetOCRBilling(observationPaymentGjuels, transmissionPaymentGjuels, t.OCRAddr)
	if err != nil {
		return err
	}

	err = t.setConfigDetails(t.OCRAddr)
	if err != nil {
		return err
	}

	return nil
}
