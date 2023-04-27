package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smartcontractkit/chainlink-starknet/ops/gauntlet"
	"os"

	caigotypes "github.com/dontpanicdao/caigo/types"

	"github.com/smartcontractkit/chainlink-starknet/integration-tests/utils"
)

var (
	ethAddressGoerli = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	nAccount         string
)

func (testState *Test) fundNodes() error {
	l := utils.GetTestLogger(testState.T)
	var nAccounts []string
	var err error
	for _, key := range testState.GetNodeKeys() {
		if key.TXKey.Data.Attributes.StarkKey == "" {
			return errors.New("stark key can't be empty")
		}
		nAccount, err = testState.Sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
		if err != nil {
			return err
		}
		if caigotypes.HexToHash(nAccount).String() != key.TXKey.Data.Attributes.AccountAddr {
			return fmt.Errorf("Deployed account with address %s not matching with node account with address %s", caigotypes.HexToHash(nAccount).String(), key.TXKey.Data.Attributes.AccountAddr)
		}
		nAccounts = append(nAccounts, nAccount)
	}

	if err != nil {
		return err
	}

	if testState.Common.Testnet {
		for _, key := range nAccounts {
			// We are not deploying in parallel here due to testnet limitations (429 too many requests)
			l.Debug().Msg(fmt.Sprintf("Funding node with address: %s", key))
			_, err = testState.Sg.TransferToken(ethAddressGoerli, key, "100000000000000000") // Transferring 1 ETH to each node
			if err != nil {
				return err
			}
		}

	} else {
		err = testState.Devnet.FundAccounts(nAccounts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (testState *Test) deployLinkToken() error {
	var err error
	testState.LinkTokenAddr, err = testState.Sg.DeployLinkTokenContract()
	if err != nil {
		return err
	}
	err = os.Setenv("LINK", testState.LinkTokenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (testState *Test) deployAccessController() error {
	var err error
	testState.AccessControllerAddr, err = testState.Sg.DeployAccessControllerContract()
	if err != nil {
		return err
	}
	err = os.Setenv("BILLING_ACCESS_CONTROLLER", testState.AccessControllerAddr)
	if err != nil {
		return err
	}
	return nil
}

func (testState *Test) setConfigDetails(ocrAddress string) error {
	cfg, err := testState.LoadOCR2Config()
	if err != nil {
		return err
	}
	var parsedConfig []byte
	parsedConfig, err = json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = testState.Sg.SetConfigDetails(string(parsedConfig), ocrAddress)
	return err
}

func (testState *Test) DeployGauntlet(minSubmissionValue int64, maxSubmissionValue int64, decimals int, name string, observationPaymentGjuels int64, transmissionPaymentGjuels int64) error {
	err := testState.Sg.InstallDependencies()
	if err != nil {
		return err
	}

	err = testState.fundNodes()
	if err != nil {
		return err
	}

	err = testState.deployLinkToken()
	if err != nil {
		return err
	}

	err = testState.deployAccessController()
	if err != nil {
		return err
	}

	testState.OCRAddr, err = testState.Sg.DeployOCR2ControllerContract(minSubmissionValue, maxSubmissionValue, decimals, name, testState.LinkTokenAddr)
	if err != nil {
		return err
	}

	testState.ProxyAddr, err = testState.Sg.DeployOCR2ProxyContract(testState.OCRAddr)
	if err != nil {
		return err
	}
	args := gauntlet.AddAccessArgs{
		ContractType: "ocr",
		Address:      testState.OCRAddr,
		Aggregator:   testState.ProxyAddr,
		User:         "",
	}

	_, err = testState.Sg.AddAccess(args)
	if err != nil {
		return err
	}

	_, err = testState.Sg.MintLinkToken(testState.LinkTokenAddr, testState.OCRAddr, "100000000000000000000")
	if err != nil {
		return err
	}
	_, err = testState.Sg.SetOCRBilling(observationPaymentGjuels, transmissionPaymentGjuels, testState.OCRAddr)
	if err != nil {
		return err
	}

	err = testState.setConfigDetails(testState.OCRAddr)
	return err
}
