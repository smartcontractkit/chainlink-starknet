package devnet

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

var (
	l2RpcUrl string
	l1RpcUrl string
	rClient  *resty.Client
)

// AutoSyncL1 auto calls /flush/ every 2 seconds to sync L1<>L2
func AutoSyncL1(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("Shutting down L1 sync")
				return
			case <-t.C:
				log.Debug().Msg("Syncing L1")
				_, err := rClient.R().Post("/postman/flush")
				if err != nil {
					log.Error().Err(err).Msg("failed to sync L1")
				}
			}
		}
	}()
}

// FundAccounts Funds provided accounts with 500000 eth each
func FundAccounts(l2AccList []string) error {
	for _, key := range l2AccList {
		_, err := rClient.R().SetBody(map[string]interface{}{
			"address": key,
			"amount":  500000,
		}).Post("/mint")
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadL1MessagingContract loads and sets up the L1 messaging contract and URL
func LoadL1MessagingContract() error {
	resp, err := rClient.R().SetBody(map[string]interface{}{
		"networkUrl": l1RpcUrl,
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}

func SetL1RpcUrl(url string) {
	l1RpcUrl = url
}

func SetL2RpcUrl(url string) {
	l2RpcUrl = url
	rClient = resty.New().SetBaseURL(l2RpcUrl)
}
