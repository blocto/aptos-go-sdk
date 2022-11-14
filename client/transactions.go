package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/portto/aptos-go-sdk/models"
)

type Transactions interface {
	GetTransactions(ctx context.Context, start, limit int, opts ...interface{}) ([]TransactionResp, error)
	SubmitTransaction(ctx context.Context, tx models.UserTransaction, opts ...interface{}) (*TransactionResp, error)
	SimulateTransaction(ctx context.Context, tx models.UserTransaction, estimateGasUnitPrice, estimateMaxGasAmount bool, opts ...interface{}) ([]TransactionResp, error)
	GetAccountTransactions(ctx context.Context, address string, start, limit int, opts ...interface{}) ([]TransactionResp, error)
	GetTransactionByHash(ctx context.Context, txHash string, opts ...interface{}) (*TransactionResp, error)
	GetTransactionByVersion(ctx context.Context, version uint64, opts ...interface{}) (*TransactionResp, error)
	EstimateGasPrice(ctx context.Context, opts ...interface{}) (uint64, error)
	WaitForTransaction(ctx context.Context, txHash string) error
}

type TransactionsImpl struct {
	Base
}

type BlockMetadataTransaction struct {
	ID                 string `json:"id"`
	Round              string `json:"round"`
	PreviousBlockVotes []bool `json:"previous_block_votes"`
	Proposer           string `json:"proposer"`
}

type TransactionResp struct {
	BlockMetadataTransaction

	Sender                  string                `json:"sender"`
	SequenceNumber          string                `json:"sequence_number"`
	Payload                 models.JSONPayload    `json:"payload"`
	MaxGasAmount            string                `json:"max_gas_amount"`
	GasUnitPrice            string                `json:"gas_unit_price"`
	ExpirationTimestampSecs string                `json:"expiration_timestamp_secs"`
	SecondarySigners        []string              `json:"secondary_signers"`
	Signature               *models.JSONSignature `json:"signature,omitempty"`

	Type                string          `json:"type"`
	Timestamp           string          `json:"timestamp"`
	Events              []models.Event  `json:"events"`
	Version             string          `json:"version"`
	Hash                string          `json:"hash"`
	StateRootHash       string          `json:"state_root_hash"`
	EventRootHash       string          `json:"event_root_hash"`
	GasUsed             string          `json:"gas_used"`
	Success             bool            `json:"success"`
	VmStatus            string          `json:"vm_status"`
	AccumulatorRootHash string          `json:"accumulator_root_hash"`
	Changes             []models.Change `json:"changes"`
}

func (impl TransactionsImpl) GetTransactions(ctx context.Context, start, limit int, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+"/v1/transactions",
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) SubmitTransaction(ctx context.Context, tx models.UserTransaction, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(ctx, http.MethodPost,
		impl.Base.Endpoint()+"/v1/transactions",
		tx, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) SimulateTransaction(ctx context.Context, tx models.UserTransaction,
	estimateGasUnitPrice, estimateMaxGasAmount bool, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(ctx, http.MethodPost,
		impl.Base.Endpoint()+"/v1/transactions/simulate",
		tx.ForSimulate(), &rspJSON, map[string]interface{}{
			"estimate_gas_unit_price": estimateGasUnitPrice,
			"estimate_max_gas_amount": estimateMaxGasAmount,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) GetAccountTransactions(ctx context.Context, address string, start, limit int, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/transactions", address),
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) GetTransactionByHash(ctx context.Context, txHash string, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/transactions/by_hash/%s", txHash),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) GetTransactionByVersion(ctx context.Context, version uint64, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/transactions/by_version/%d", version),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) EstimateGasPrice(ctx context.Context, opts ...interface{}) (uint64, error) {
	type response struct {
		GasEstimate uint64 `json:"gas_estimate"`
	}
	var rspJSON response
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+"/v1/estimate_gas_price",
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return 0, err
	}

	return rspJSON.GasEstimate, nil
}

const (
	// the maximum retry count of WaitForTransaction
	maxRetryCount = 10
)

func (impl TransactionsImpl) WaitForTransaction(ctx context.Context, txHash string) error {
	var isPending bool = true
	var count int
	for isPending && count < maxRetryCount {
		tx, err := impl.GetTransactionByHash(ctx, txHash)
		isPending = (err != nil || tx.Type == "pending_transaction")
		if isPending {
			time.Sleep(1 * time.Second)
			count += 1
		}
	}

	if isPending {
		return fmt.Errorf("transaction %s timed out", txHash)
	}
	return nil
}
