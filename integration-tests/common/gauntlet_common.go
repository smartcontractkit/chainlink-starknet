package common

import (
	"encoding/json"
	"errors"
	"fmt"
	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-testing-framework/client"
	"os"
	"sync"
)

var (
	ethAddressGoerli = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
	nAccount         string
)

func (t *Test) fundNodes() error {
	var nAccounts []string
	errChannel := make(chan error)
	wgDone := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(len(t.GetNodeKeys()))
	for _, key := range t.GetNodeKeys() {
		go func(key client.NodeKeysBundle) {
			if key.TXKey.Data.Attributes.StarkKey == "" {
				errChannel <- errors.New("stark key can't be empty")
			}
			nAccount, err = t.Sg.DeployAccountContract(100, key.TXKey.Data.Attributes.StarkKey)
			if err != nil {
				errChannel <- err
			}
			if caigotypes.HexToHash(nAccount).String() != key.TXKey.Data.Attributes.AccountAddr {
				errChannel <- errors.New(fmt.Sprintf("Deployed account with address %s not matching with node account with address %s", caigotypes.HexToHash(nAccount).String(), key.TXKey.Data.Attributes.AccountAddr))
			}
			nAccounts = append(nAccounts, nAccount)
			defer wg.Done()
		}(key)
	}
	wg.Wait()
	close(wgDone)
	select {
	case <-wgDone:
		break
	case err = <-errChannel:
		close(errChannel)
		return err
	}
	errChannel = make(chan error)
	wgDone = make(chan bool)
	if t.Testnet {
		wg.Add(len(nAccounts))
		for _, key := range nAccounts {
			go func(key string) {
				log.Debug().Msg(fmt.Sprintf("Funding node with address: %s", key))
				_, err = t.Sg.TransferToken(ethAddressGoerli, key, "10000000000000000") // Transferring 0.1 ETH to each node
				if err != nil {
					errChannel <- err
				}
				defer wg.Done()
			}(key)

		}
	} else {
		err = t.Devnet.FundAccounts(nAccounts)
		if err != nil {
			return err
		}
	}
	wg.Wait()
	close(wgDone)
	select {
	case <-wgDone:
		break
	case err = <-errChannel:
		close(errChannel)
		return err
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
	_, err = t.Sg.AddAccess(t.OCRAddr, t.ProxyAddr)
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
