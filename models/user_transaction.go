package models

import (
	"strconv"

	"github.com/the729/lcs"
)

type UserTransaction struct {
	RawTransaction
	SecondarySigners []AccountAddress
	Signature        *Signature `lcs:"-"`
}

func (tx UserTransaction) GetRawTransactionWithData() RawTransactionWithData {
	return MultiAgent{
		RawTransaction:   tx.RawTransaction,
		SecondarySigners: tx.SecondarySigners,
	}
}

type RawTransaction struct {
	Sender                  AccountAddress
	SequenceNumber          uint64
	Payload                 TransactionPayload
	MaxGasAmount            uint64
	GasUnitPrice            uint64
	ExpirationTimestampSecs uint64
	ChainID                 uint8
}

type RawTransactionWithData interface{}

var _ = lcs.RegisterEnum(
	(*RawTransactionWithData)(nil),
	MultiAgent{},
)

type MultiAgent struct {
	RawTransaction
	SecondarySigners []AccountAddress
}

func (tx UserTransaction) ToRequest() UserTransactionRequest {
	req := UserTransactionRequest{
		Sender:                  tx.Sender.PrefixZeroTrimmedHex(),
		SequenceNumber:          strconv.FormatUint(tx.SequenceNumber, 10),
		Payload:                 tx.Payload.ToJSON(),
		MaxGasAmount:            strconv.FormatUint(tx.MaxGasAmount, 10),
		GasUnitPrice:            strconv.FormatUint(tx.GasUnitPrice, 10),
		ExpirationTimestampSecs: strconv.FormatUint(tx.ExpirationTimestampSecs, 10),
		Signature:               tx.Signature,
	}

	if len(tx.SecondarySigners) > 0 {
		req.SecondarySigners = make([]string, len(tx.SecondarySigners))
		for i, signer := range tx.SecondarySigners {
			req.SecondarySigners[i] = signer.PrefixZeroTrimmedHex()
		}
	}

	return req
}

type UserTransactionRequest struct {
	Sender                  string      `json:"sender"`
	SequenceNumber          string      `json:"sequence_number"`
	Payload                 JSONPayload `json:"payload"`
	MaxGasAmount            string      `json:"max_gas_amount"`
	GasUnitPrice            string      `json:"gas_unit_price"`
	ExpirationTimestampSecs string      `json:"expiration_timestamp_secs"`
	SecondarySigners        []string    `json:"secondary_signers,omitempty"`
	*Signature              `json:"signature,omitempty"`
}
