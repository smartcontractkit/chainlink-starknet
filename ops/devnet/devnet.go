package devnet

import (
	"context"
	"fmt"
	"strings"
	"time"

	starknetutils "github.com/NethermindEth/starknet.go/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type StarknetDevnetClient struct {
	ctx      context.Context
	cancel   context.CancelFunc
	client   *resty.Client
	dumpPath string
}

func (devnet *StarknetDevnetClient) NewStarknetDevnetClient(rpcUrl string, dumpPath string) *StarknetDevnetClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &StarknetDevnetClient{
		ctx:      ctx,
		cancel:   cancel,
		client:   resty.New().SetBaseURL(rpcUrl),
		dumpPath: dumpPath,
	}
}

// AutoSyncL1 auto calls /flush/ every 2 seconds to sync L1<>L2
func (devnet *StarknetDevnetClient) AutoSyncL1() {
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
func (devnet *StarknetDevnetClient) AutoDumpState() {
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
func (devnet *StarknetDevnetClient) AutoLoadState(client *ocr2.Client, ocrAddress string) {
	addr, _ := starknetutils.HexToFelt(ocrAddress)
	t := time.NewTicker(15 * time.Second)
	go func() {
		for {
			select {
			case <-devnet.ctx.Done():
				log.Debug().Msg("Shutting down devnet dump")
				return
			case <-t.C:
				log.Debug().Msg("Checking for devnet OCR contract errors")

				_, err := client.LatestTransmissionDetails(devnet.ctx, addr)
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
func (devnet *StarknetDevnetClient) FundAccounts(l2AccList []string) error {
	for _, key := range l2AccList {
		res, err := devnet.client.R().SetBody(map[string]any{
			"address": key,
			"amount":  9000000000000000000000.0,
		}).Post("/mint")
		if err != nil {
			return err
		}
		log.Info().Msg(fmt.Sprintf("Funding account: %s", string(res.Body())))
	}
	return nil
}

// LoadL1MessagingContract loads and sets up the L1 messaging contract and URL
func (devnet *StarknetDevnetClient) LoadL1MessagingContract(l1RpcUrl string) error {
	resp, err := devnet.client.R().SetBody(map[string]any{
		"networkUrl": l1RpcUrl,
	}).Post("/postman/load_l1_messaging_contract")
	if err != nil {
		return err
	}
	log.Warn().Interface("Response", resp.String()).Msg("Set up L1 messaging contract")
	return nil
}
