package client

import (
	"fmt"
	"net/http"

	"github.com/portto/aptos-go-sdk/models"
)

type Transactions interface {
	GetTransactions(start, limit int, opts ...interface{}) ([]TransactionResp, error)
	SubmitTransaction(tx models.UserTransaction, opts ...interface{}) (*TransactionResp, error)
	SimulateTransaction(tx models.UserTransaction, opts ...interface{}) ([]TransactionResp, error)
	GetAccountTransactions(address string, start, limit int, opts ...interface{}) ([]TransactionResp, error)
	GetTransactionByHash(txHash string, opts ...interface{}) (*TransactionResp, error)
	GetTransactionByVersion(version uint64, opts ...interface{}) (*TransactionResp, error)
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

func (impl TransactionsImpl) GetTransactions(start, limit int, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(http.MethodGet,
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

func (impl TransactionsImpl) SubmitTransaction(tx models.UserTransaction, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(http.MethodPost,
		impl.Base.Endpoint()+"/v1/transactions",
		tx, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) SimulateTransaction(tx models.UserTransaction, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(http.MethodPost,
		impl.Base.Endpoint()+"/v1/transactions/simulate",
		tx.ForSimulate(), &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) GetAccountTransactions(address string, start, limit int, opts ...interface{}) ([]TransactionResp, error) {
	var rspJSON []TransactionResp
	err := request(http.MethodGet,
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

func (impl TransactionsImpl) GetTransactionByHash(txHash string, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/transactions/by_hash/%s", txHash),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) GetTransactionByVersion(version uint64, opts ...interface{}) (*TransactionResp, error) {
	var rspJSON TransactionResp
	err := request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/transactions/by_version/%d", version),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
