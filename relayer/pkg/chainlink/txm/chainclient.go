package txm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/core"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

type starkChainClient struct {
	client *utils.LazyLoad[starknet.ReaderWriter]
}

var _ core.ChainClient[keys.Key, types.Transaction, *big.Int, *types.Felt] = starkChainClient{}

func (cc starkChainClient) GetNonce(ctx context.Context, k keys.Key, tx types.Transaction) (nonce *big.Int, err error) {
	client, err := cc.client.Get()
	if err != nil {
		return nil, err
	}

	// allow custom passed nonce
	var ok bool
	if tx.Nonce != "" {
		nonce, ok = new(big.Int).SetString(tx.Nonce, 0) // use base 0 to dynamically determine base
		if !ok {
			return nonce, errors.Errorf("failed to decode custom nonce: %s", tx.Nonce)
		}
	} else {
		// TODO: investigate if nonce management is needed (nonce is requested queried by the sdk for now)
		// get nonce
		nonce, err = client.AccountNonce(ctx, tx.SenderAddress)
		if err != nil {
			return nonce, errors.Errorf("failed to fetch nonce: %s", err)
		}
	}

	return nonce, nil
}

func (cc starkChainClient) EstimateTx(ctx context.Context, k keys.Key, tx types.Transaction, nonce *big.Int) (*types.Felt, error) {
	client, err := cc.client.Get()
	if err != nil {
		return nil, err
	}

	// parse key to expected format
	privKeyBytes := k.Raw()
	privKey := caigo.BigToHex(caigo.BytesToBig(privKeyBytes))

	// create new account
	account, err := caigo.NewAccount(privKey, tx.SenderAddress, client) // priv key not needed to simulate
	if err != nil {
		return nil, errors.Errorf("%s: %s", AccountCreationErr, err)
	}

	// set nonce
	details := caigo.ExecuteDetails{
		Nonce: nonce,
	}

	// allow custom passed max fee
	var fee *big.Int
	var ok bool
	if tx.MaxFee != "" {
		fee, ok = new(big.Int).SetString(tx.MaxFee, 0) // use base 0 to dynamically determine base
		if !ok {
			return nil, errors.Errorf("failed to decode custom max fee: %s", tx.MaxFee)
		}
	} else {
		// nonce management + fee estimator go together (otherwise too high of nonce will cause estimate to fail)
		// get fee for txm
		feeData, err := account.EstimateFee(ctx, []types.Transaction{tx}, details)
		if err != nil {
			return nil, errors.Errorf("failed to estimate fee: %s", err)
		}
		fee = new(big.Int).SetUint64((feeData.OverallFee * FEE_MARGIN) / 100)
	}
	return types.BigToFelt(fee), nil
}

func (cc starkChainClient) SendTx(ctx context.Context, k keys.Key, tx types.Transaction, nonce *big.Int, maxFee *types.Felt) (txhash string, err error) {
	client, err := cc.client.Get()
	if err != nil {
		return txhash, err
	}

	// parse key to expected format
	privKeyBytes := k.Raw()
	privKey := caigo.BigToHex(caigo.BytesToBig(privKeyBytes))

	// create new account
	account, err := caigo.NewAccount(privKey, tx.SenderAddress, client)
	if err != nil {
		return txhash, errors.Errorf("%s: %s", AccountCreationErr, err)
	}

	// set nonce
	details := caigo.ExecuteDetails{
		Nonce:  nonce,
		MaxFee: maxFee,
	}

	// transmit txs
	res, err := account.Execute(ctx, []types.Transaction{tx}, details)
	if err != nil {
		return txhash, errors.Errorf("failed to execute tx: %s", err)
	}

	// handle nil pointer
	if res == nil {
		return txhash, errors.New("execute response and error are nil")
	}

	return res.TransactionHash, nil
}

func (cc starkChainClient) TxStatus(ctx context.Context, txhash string) (status core.Status, errStr string, err error) {
	client, err := cc.client.Get()
	if err != nil {
		return status, errStr, err
	}

	starknetStatus, err := client.TransactionStatus(ctx, gateway.TransactionStatusOptions{
		TransactionHash: txhash,
	})
	if err != nil {
		return status, errStr, fmt.Errorf("failed to fetch receipt, hash: %s", txhash)
	}

	if starknetStatus.TxStatus == types.ACCEPTED_ON_L1.String() || starknetStatus.TxStatus == types.ACCEPTED_ON_L2.String() {
		status = core.CONFIRMED
	} else if starknetStatus.TxStatus == types.REJECTED.String() {
		status = core.ERRORED
		errStr = starknetStatus.TxFailureReason.ErrorMessage
	}
	return status, errStr, nil
}

func (cc starkChainClient) IsFatalError(errStr string) bool {
	// Fatal chain specific conditions
	// TODO: verify if this is true
	return strings.Contains(errStr, "Error in the called contract") || // https://github.com/starkware-libs/cairo-lang/blob/master/src/starkware/starknet/business_logic/execution/execute_entry_point.py#L237
		strings.Contains(errStr, "No contract at the provided address") ||
		strings.Contains(errStr, AccountCreationErr) // err only occurs if private key is invalid: https://github.com/dontpanicdao/caigo/blob/main/account.go#L42
}