package models

import (
	"strconv"
	"strings"

	"github.com/the729/lcs"
)

type TransactionEnum interface{}

var _ = lcs.RegisterEnum(
	(*TransactionEnum)(nil),
	UserTransaction{},
)

type UserTransaction struct {
	RawTransaction
	Authenticator    TransactionAuthenticator
	SecondarySigners []AccountAddress `lcs:"-"`
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
	}

	if tx.Authenticator != nil {
		req.Signature = tx.Authenticator.ToJSONSignature()
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
	Sender                  string         `json:"sender"`
	SequenceNumber          string         `json:"sequence_number"`
	Payload                 JSONPayload    `json:"payload"`
	MaxGasAmount            string         `json:"max_gas_amount"`
	GasUnitPrice            string         `json:"gas_unit_price"`
	ExpirationTimestampSecs string         `json:"expiration_timestamp_secs"`
	SecondarySigners        []string       `json:"secondary_signers,omitempty"`
	Signature               *JSONSignature `json:"signature,omitempty"`
}

func (tx UserTransactionRequest) ForSimulate() UserTransactionRequest {
	zeroSig := "0x" + strings.Repeat("00", 64)
	tx.Signature = &JSONSignature{
		Type:                  tx.Signature.Type,
		ED25519Signature:      tx.Signature.ED25519Signature,
		MultiED25519Signature: tx.Signature.MultiED25519Signature,
		MultiAgentSignature:   tx.Signature.MultiAgentSignature,
	}

	if tx.Signature != nil {
		if len(tx.Signature.Signature) > 0 {
			tx.Signature.Signature = zeroSig
		}

		newSignatures := make([]string, len(tx.Signature.Signatures))
		for i, sig := range tx.Signature.Signatures {
			if len(sig) > 0 {
				newSignatures[i] = zeroSig
			}
		}
		tx.Signature.Signatures = newSignatures

		if len(tx.Signature.Sender.Signature) > 0 {
			tx.Signature.Sender.Signature = zeroSig
		}

		newSignatures = make([]string, len(tx.Signature.Sender.Signatures))
		for i, sig := range tx.Signature.Sender.Signatures {
			if len(sig) > 0 {
				newSignatures[i] = zeroSig
			}
		}
		tx.Signature.Sender.Signatures = newSignatures

		newSecondarySigners := make([]JSONSigner, len(tx.Signature.SecondarySigners))
		for i, signer := range tx.Signature.SecondarySigners {
			newSecondarySigners[i].Type = signer.Type
			newSecondarySigners[i].PublicKey = signer.PublicKey
			newSecondarySigners[i].PublicKeys = signer.PublicKeys
			newSecondarySigners[i].Threshold = signer.Threshold
			newSecondarySigners[i].Bitmap = signer.Bitmap
			if len(signer.Signature) > 0 {
				newSecondarySigners[i].Signature = zeroSig
			}

			newSignatures := make([]string, len(signer.Signatures))
			for ii, sig := range signer.Signatures {
				if len(sig) > 0 {
					newSignatures[ii] = zeroSig
				}
			}
			newSecondarySigners[i].Signatures = newSignatures
		}
		tx.Signature.SecondarySigners = newSecondarySigners
	}

	return tx
}
