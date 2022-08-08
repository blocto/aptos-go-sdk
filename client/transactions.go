package client

import (
	"fmt"
	"net/http"
)

type Transactions interface {
	GetTransactions(start, limit int) ([]Transaction, error)
	SubmitTransaction(tx Transaction) (*Transaction, error)
	SimulateTransaction(tx Transaction) ([]Transaction, error)
	GetAccountTransactions(address string, start, limit int) ([]Transaction, error)
	GetTransaction(txHash string) (*Transaction, error)
	CreateTransactionSigningMessage(tx Transaction) (*SigningMessage, error)
}

type TransactionsImpl struct {
	Base
}

type ED25519Signature struct {
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

type MultiED25519Signature struct {
	PublicKeys []string `json:"public_keys"`
	Signatures []string `json:"signatures"`
	Threshold  int      `json:"threshold"`
	Bitmap     string   `json:"bitmap"`
}

type MultiAgentSignature struct {
	Sender struct {
		Type string `json:"type"`
		ED25519Signature
		MultiED25519Signature
	} `json:"sender"`
	SecondarySignerAddresses []string `json:"secondary_signer_addresses"`
	SecondarySigners         []struct {
		Type string `json:"type"`
		ED25519Signature
		MultiED25519Signature
	} `json:"secondary_signers"`
}

type ScriptFunctionPayload struct {
	Function string `json:"function"`
}

type WriteSetPayload struct {
	DirectWriteSet
	ScriptWriteSet
}

type ScriptPayload struct {
	Code Code `json:"code"`
}

type ModuleBundlePayload struct {
	Modules []Code `json:"modules"`
}

type Code struct {
	Bytecode string      `json:"bytecode"`
	ABI      interface{} `json:"abi"`
}

type ScriptWriteSet struct {
	ExecuteAs string `json:"execute_as"`
	Script    struct {
		Code          Code          `json:"code"`
		TypeArguments []string      `json:"type_arguments"`
		Arguments     []interface{} `json:"arguments"`
	} `json:"script"`
}

type DirectWriteSet struct {
	Changes []Change `json:"changes"`
	Events  []Event  `json:"events"`
}

type Change struct {
	Type         string `json:"type"`
	StateKeyHash string `json:"state_key_hash"`
	Address      string `json:"address"`
	Module       string `json:"module"`
	Resource     string `json:"resource"`
	Data         struct {
		Handle   string                 `json:"handle"`
		Key      string                 `json:"key"`
		Value    string                 `json:"value"`
		Bytecode string                 `json:"bytecode"`
		ABI      interface{}            `json:"abi"`
		Type     string                 `json:"type"`
		Data     map[string]interface{} `json:"data"`
	} `json:"data"`
}

type UserTransaction struct {
	Sender                  string `json:"sender"`
	SequenceNumber          string `json:"sequence_number"`
	MaxGasAmount            string `json:"max_gas_amount"`
	GasUnitPrice            string `json:"gas_unit_price"`
	GasCurrencyCode         string `json:"gas_currency_code"`
	ExpirationTimestampSecs string `json:"expiration_timestamp_secs"`
	Signature               *struct {
		Type string `json:"type"`
		MultiED25519Signature
		ED25519Signature
		MultiAgentSignature
	} `json:"signature,omitempty"`
}

type BlockMetadataTransaction struct {
	ID                 string   `json:"id"`
	Round              string   `json:"round"`
	PreviousBlockVotes []string `json:"previous_block_votes"`
	Proposer           string   `json:"proposer"`
}

type Transaction struct {
	BlockMetadataTransaction
	UserTransaction

	Type      string  `json:"type"`
	Timestamp string  `json:"timestamp"`
	Events    []Event `json:"events"`
	Payload   struct {
		Type          string        `json:"type"`
		TypeArguments []string      `json:"type_arguments"`
		Arguments     []interface{} `json:"arguments"`
		ScriptFunctionPayload
		ScriptPayload
		WriteSetPayload
		ModuleBundlePayload
	} `json:"payload"`
	Version             string   `json:"version"`
	Hash                string   `json:"hash"`
	StateRootHash       string   `json:"state_root_hash"`
	EventRootHash       string   `json:"event_root_hash"`
	GasUsed             string   `json:"gas_used"`
	Success             bool     `json:"success"`
	VmStatus            string   `json:"vm_status"`
	AccumulatorRootHash string   `json:"accumulator_root_hash"`
	Changes             []Change `json:"changes"`
}

func (impl TransactionsImpl) GetTransactions(start, limit int) ([]Transaction, error) {
	var rspJSON []Transaction
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/transactions"),
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		})
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) SubmitTransaction(tx Transaction) (*Transaction, error) {
	var rspJSON Transaction
	err := Request(http.MethodPost,
		impl.Base.Endpoint()+"/transactions",
		tx, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl TransactionsImpl) SimulateTransaction(tx Transaction) ([]Transaction, error) {
	var rspJSON []Transaction
	err := Request(http.MethodPost,
		impl.Base.Endpoint()+"/transactions/simulate",
		tx, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) GetAccountTransactions(address string, start, limit int) ([]Transaction, error) {
	var rspJSON []Transaction
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/transactions", address),
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		})
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl TransactionsImpl) GetTransaction(txHash string) (*Transaction, error) {
	var rspJSON Transaction
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/transactions/%s", txHash),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type SigningMessage struct {
	Message string `json:"message"`
}

func (impl TransactionsImpl) CreateTransactionSigningMessage(tx Transaction) (*SigningMessage, error) {
	var rspJSON SigningMessage
	err := Request(http.MethodPost,
		impl.Base.Endpoint()+"/transactions/signing_message",
		tx, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
