package devnet

import (
	"context"
	"fmt"
	"strings"
	"time"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type StarkNetDevnetClient struct {
	ctx      context.Context
	cancel   context.CancelFunc
	client   *resty.Client
	dumpPath string
}

func (devnet *StarkNetDevnetClient) NewStarkNetDevnetClient(rpcUrl string, dumpPath string) *StarkNetDevnetClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &StarkNetDevnetClient{
		ctx:      ctx,
		cancel:   cancel,
		client:   resty.New().SetBaseURL(rpcUrl),
		dumpPath: dumpPath,
	}
}

// AutoSyncL1 auto calls /flush/ every 2 seconds to sync L1<>L2
func (devnet *StarkNetDevnetClient) AutoSyncL1() {
	t := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-devnet.ctx.Done():
				log.Debug().Msg("Shutting down L1 sync")
				return
			case <-t.C:
				log.Debug().Msg("Syncing L1")
				_, err := devnet.client.R().Post("/postman/flush")
				if err != nil {
					log.Error().Err(err).Msg("failed to sync L1")
				}
			}
		}
	}()
}

// AutoDumpState dumps devnet state every 10 sec
func (devnet *StarkNetDevnetClient) AutoDumpState() {
	t := time.NewTicker(20 * time.Minute)
	go func() {
		for {
			select {
			case <-devnet.ctx.Done():
				log.Debug().Msg("Shutting down devnet dump")
				return
			case <-t.C:
				log.Debug().Msg("Dumping state")
				_, err := devnet.client.R().SetBody(map[string]any{
					"path": devnet.dumpPath,
				}).Post("/dump")
				if err != nil {
					log.Error().Err(err).Msg("Failed to dump devnet state")
				}
			}
		}
	}()
}

// AutoLoadState auto loads last saved devnet state on contract not found
func (devnet *StarkNetDevnetClient) AutoLoadState(client *ocr2.Client, ocrAddress string) {
	t := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-devnet.ctx.Done():
				log.Debug().Msg("Shutting down devnet dump")
				return
			case <-t.C:
				log.Debug().Msg("Checking for devnet OCR contract errors")
				_, err := client.LatestTransmissionDetails(devnet.ctx, caigotypes.HexToHash(ocrAddress))
				if err != nil && strings.Contains(err.Error(), "is not deployed") {
					_, err = devnet.client.R().SetBody(map[string]any{
						"path": devnet.dumpPath,
					}).Post("/load")
					if err != nil {
						log.Error().Err(err).Msg("Failed to dump devnet state")
					}
				}

			}
		}
	}()
}

// FundAccounts Funds provided accounts with 100 eth each
func (devnet *StarkNetDevnetClient) FundAccounts(l2AccList []string) error {
	for _, key := range l2AccList {
		res, err := devnet.client.R().SetBody(map[string]any{
			"address": key,
			"amount":  1e21,
		}).Post("/mint")
		if err != nil {
			return err
		}
		log.Info().Msg(fmt.Sprintf("Funding account: %s", string(res.Body())))
	}
	return nil
}

// LoadL1MessagingContract loads and sets up the L1 messaging contract and URL
func (devnet *StarkNetDevnetClient) LoadL1MessagingContract(l1RpcUrl string) error {
	resp, err := devnet.client.R().SetBody(map[string]any{
		"networkUrl": l1RpcUrl,
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}
