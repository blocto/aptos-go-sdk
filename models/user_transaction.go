package models

import (
	"crypto/ed25519"
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

func (tx UserTransaction) ForSimulate() UserTransaction {
	var zeroSig Signature = make([]byte, ed25519.SignatureSize)

	switch auth := tx.Authenticator.(type) {
	case TransactionAuthenticatorEd25519:
		auth.Signature = zeroSig
		tx.Authenticator = auth
	case TransactionAuthenticatorMultiEd25519:
		zeroSignatures := make([]Signature, len(auth.Signatures))
		for i := range zeroSignatures {
			zeroSignatures[i] = zeroSig
		}
		auth.Signatures = zeroSignatures
		tx.Authenticator = auth.SetBytes()
	case TransactionAuthenticatorMultiAgent:
		switch sender := auth.Sender.(type) {
		case AccountAuthenticatorEd25519:
			sender.Signature = zeroSig
			auth.Sender = sender
		case AccountAuthenticatorMultiEd25519:
			zeroSignatures := make([]Signature, len(sender.Signatures))
			for i := range zeroSignatures {
				zeroSignatures[i] = zeroSig
			}
			sender.Signatures = zeroSignatures
			auth.Sender = sender.SetBytes()
		}

		for i, signer := range auth.SecondarySigners {
			switch signer := signer.(type) {
			case AccountAuthenticatorEd25519:
				signer.Signature = zeroSig
				auth.SecondarySigners[i] = signer
			case AccountAuthenticatorMultiEd25519:
				zeroSignatures := make([]Signature, len(signer.Signatures))
				for i := range zeroSignatures {
					zeroSignatures[i] = zeroSig
				}
				signer.Signatures = zeroSignatures
				auth.SecondarySigners[i] = signer
			}
		}

		tx.Authenticator = auth
	}

	return tx
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
