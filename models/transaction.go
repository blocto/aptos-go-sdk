package models

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/the729/lcs"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

type Transaction struct {
	err error

	// signing message for verification
	signingMessage []byte
	hash           string

	UserTransaction
}

func (t *Transaction) SetChainID(chainID uint8) *Transaction {
	if t.hasError() {
		return t
	}

	t.ChainID = chainID
	return t
}

func (t *Transaction) SetSender(sender string) *Transaction {
	if t.hasError() {
		return t
	}

	addr, err := HexToAccountAddress(sender)
	if err != nil {
		t.err = err
		return t
	}

	t.Sender = addr
	return t
}

func (t *Transaction) SetSequenceNumber(seq interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch seq := seq.(type) {
	case uint64:
		t.SequenceNumber = seq
	case string:
		s, ok := new(big.Int).SetString(seq, 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.SequenceNumber = s.Uint64()
	default:
		t.err = fmt.Errorf("unexpected type: %T", seq)
		return t
	}

	return t
}

func (t *Transaction) SetMaxGasAmount(maxGasAmount interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch maxGasAmount := maxGasAmount.(type) {
	case uint64:
		t.MaxGasAmount = maxGasAmount
	case string:
		s, ok := new(big.Int).SetString(maxGasAmount, 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.MaxGasAmount = s.Uint64()
	default:
		t.err = fmt.Errorf("unexpected type: %T", maxGasAmount)
		return t
	}

	return t
}

func (t *Transaction) SetGasUnitPrice(gasUnitPrice interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch gasUnitPrice := gasUnitPrice.(type) {
	case uint64:
		t.GasUnitPrice = gasUnitPrice
	case string:
		s, ok := new(big.Int).SetString(gasUnitPrice, 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.GasUnitPrice = s.Uint64()
	default:
		t.err = fmt.Errorf("unexpected type: %T", gasUnitPrice)
		return t
	}

	return t
}

func (t *Transaction) SetExpirationTimestampSecs(expirationTimestampSecs interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch secs := expirationTimestampSecs.(type) {
	case uint64:
		t.ExpirationTimestampSecs = secs
	case string:
		s, ok := new(big.Int).SetString(secs, 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.ExpirationTimestampSecs = s.Uint64()
	default:
		t.err = fmt.Errorf("unexpected type: %T", expirationTimestampSecs)
		return t
	}
	return t
}

func (t *Transaction) SetPayload(payload TransactionPayload) *Transaction {
	if t.hasError() {
		return t
	}

	switch payload := payload.(type) {
	case ScriptPayload:
		if payload.TypeArguments == nil {
			payload.TypeArguments = make([]TypeTag, 0)
		}

		if payload.Arguments == nil {
			payload.Arguments = make([]TransactionArgument, 0)
		}

		t.Payload = payload
	case ModuleBundlePayload:
		t.Payload = payload
	case EntryFunctionPayload:
		if payload.TypeArguments == nil {
			payload.TypeArguments = make([]TypeTag, 0)
		}

		payload.ArgumentsBCS = make([][]byte, len(payload.Arguments))
		for i, arg := range payload.Arguments {
			switch arg := arg.(type) {
			case AccountAddress:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case [32]byte:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case []byte:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case string:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case uint64:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case uint8:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			case bool:
				payload.ArgumentsBCS[i], t.err = lcs.Marshal(&arg)
			}
			if t.err != nil {
				t.err = fmt.Errorf("marshal arguments[%d] %v: %v", i, arg, t.err)
				return t
			}
		}
		t.Payload = payload
	default:
		t.err = fmt.Errorf("unexpected payload type %T", payload)
		return t
	}

	return t
}

func (t *Transaction) SetAuthenticator(txAuth TransactionAuthenticator) *Transaction {
	if t.hasError() {
		return t
	}

	switch txAuth := txAuth.(type) {
	case TransactionAuthenticatorEd25519:
		if !ed25519.Verify(txAuth.PublicKey, t.signingMessage, txAuth.Signature) {
			t.err = errors.New("ed25519.Verify failed")
			return t
		}
		t.Authenticator = txAuth
	case TransactionAuthenticatorMultiEd25519:
		if err := t.validateMultiEd25519(txAuth); err != nil {
			t.err = err
			return t
		}

		t.Authenticator = txAuth.SetBytes()
	case TransactionAuthenticatorMultiAgent:
		if err := t.validateMultiAgent(txAuth); err != nil {
			t.err = err
			return t
		}

		switch sender := txAuth.Sender.(type) {
		case AccountAuthenticatorEd25519:
		case AccountAuthenticatorMultiEd25519:
			txAuth.Sender = sender.SetBytes()
		}

		for i, signer := range txAuth.SecondarySigners {
			switch signer := signer.(type) {
			case AccountAuthenticatorEd25519:
			case AccountAuthenticatorMultiEd25519:
				txAuth.SecondarySigners[i] = signer.SetBytes()
			}
		}
		t.Authenticator = txAuth
	default:
		t.err = fmt.Errorf("unexpected signature type %T", txAuth)
		return t
	}

	return t
}

func (t *Transaction) SetSecondarySigners(secondarySigners []AccountAddress) *Transaction {
	if t.hasError() {
		return t
	}

	t.SecondarySigners = secondarySigners
	return t
}

func (t Transaction) validateMultiEd25519(txAuth TransactionAuthenticatorMultiEd25519) error {
	if len(txAuth.Signatures) < int(txAuth.Threshold) {
		return fmt.Errorf("signatures size(%d) must >= threshold(%d)",
			len(txAuth.Signatures), txAuth.Threshold)
	}

	var sigIndex int
	bitmapBigInt := new(big.Int).SetBytes(txAuth.Bitmap[:])
	for i := 0; i < 32; i++ {
		if bitmapBigInt.Bit(31-i) == 1 {
			if i > len(txAuth.PublicKeys)-1 {
				return fmt.Errorf("bitmap %b and public keys length %d not match",
					txAuth.Bitmap, len(txAuth.PublicKeys))
			}

			if !ed25519.Verify(txAuth.PublicKeys[i], t.signingMessage, txAuth.Signatures[sigIndex]) {
				return fmt.Errorf("ed25519.Verify failed")
			}
			sigIndex++
		}
	}

	if sigIndex != len(txAuth.Signatures) {
		return fmt.Errorf("does not have enough signatures: %d vs %d",
			sigIndex, len(txAuth.Signatures))
	}
	return nil
}

func (t Transaction) validateMultiAgent(txAuth TransactionAuthenticatorMultiAgent) error {
	switch sender := txAuth.Sender.(type) {
	case AccountAuthenticatorEd25519:
		if !ed25519.Verify(sender.PublicKey, t.signingMessage, sender.Signature) {
			return errors.New("ed25519.Verify failed")
		}
	case AccountAuthenticatorMultiEd25519:
		if err := t.validateMultiEd25519(TransactionAuthenticatorMultiEd25519(sender)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected sender type %T", sender)
	}

	if len(txAuth.SecondarySignerAddresses) != len(txAuth.SecondarySigners) {
		return fmt.Errorf("incorrect agent signatures size: %d vs %d",
			len(txAuth.SecondarySigners), len(txAuth.SecondarySigners))
	}

	for _, signer := range txAuth.SecondarySigners {
		switch signer := signer.(type) {
		case AccountAuthenticatorEd25519:
			if !ed25519.Verify(signer.PublicKey, t.signingMessage, signer.Signature) {
				return errors.New("ed25519.Verify failed")
			}
		case AccountAuthenticatorMultiEd25519:
			if err := t.validateMultiEd25519(TransactionAuthenticatorMultiEd25519(signer)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected signer type %T", signer)
		}
	}

	return nil
}

var TransactionSalt = sha3.Sum256([]byte("APTOS::Transaction"))

func (t *Transaction) GetHash() (string, error) {
	if t.hash != "" {
		return t.hash, nil
	}

	var tx TransactionEnum = t.UserTransaction
	bcsBytes, err := lcs.Marshal(&tx)
	if err != nil {
		return "", err
	}

	hash := sha3.Sum256(append(TransactionSalt[:], bcsBytes...))

	t.hash = "0x" + hex.EncodeToString(hash[:])
	return t.hash, nil
}

// "0xb5e97db07fa0bd0e5598aa3643a9bc6f6693bddc1a9fec9e674a461eaa00b193"
var RawTransactionSalt = sha3.Sum256([]byte("APTOS::RawTransaction"))

// "0x5efa3c4f02f83a0f4b2d69fc95c607cc02825cc4e7be536ef0992df050d9e67c"
var RawTransactionWithDataSalt = sha3.Sum256([]byte("APTOS::RawTransactionWithData"))

func (t *Transaction) GetSigningMessage() ([]byte, error) {
	if t.signingMessage != nil {
		return t.signingMessage, nil
	}

	// MultiAgentRawTransaction
	if len(t.SecondarySigners) > 0 {
		rawTransactionWithData := t.GetRawTransactionWithData()
		bcsBytes, err := lcs.Marshal(&rawTransactionWithData)
		if err != nil {
			return nil, err
		}

		t.signingMessage = append(RawTransactionWithDataSalt[:], bcsBytes...)
	} else {
		bcsBytes, err := lcs.Marshal(t.RawTransaction)
		if err != nil {
			return nil, err
		}

		t.signingMessage = append(RawTransactionSalt[:], bcsBytes...)
	}

	return t.signingMessage, nil
}

func (t *Transaction) DecodeFromSigningMessageHex(s string) error {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	bcsBytes, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("hex.DecodeFromHex error: %v", err)
	}

	if len(bcsBytes) < 32 {
		return fmt.Errorf("incorrect prefix len: %d", len(s))
	}

	switch prefix := hex.EncodeToString(bcsBytes[:32]); prefix {
	case hex.EncodeToString(RawTransactionWithDataSalt[:]):
		var rawTransactionWithData RawTransactionWithData = MultiAgent{}
		if err := lcs.Unmarshal(bcsBytes[32:], &rawTransactionWithData); err != nil {
			return fmt.Errorf("MultiAgent lcs.Unmarshal error: %v", err)
		}

		t.UserTransaction.RawTransaction = rawTransactionWithData.(MultiAgent).RawTransaction
		t.UserTransaction.SecondarySigners = rawTransactionWithData.(MultiAgent).SecondarySigners
	case hex.EncodeToString(RawTransactionSalt[:]):
		if err := lcs.Unmarshal(bcsBytes[32:], &t.UserTransaction.RawTransaction); err != nil {
			return fmt.Errorf("RawTransaction lcs.Unmarshal error: %v", err)
		}
	default:
		return fmt.Errorf("unexpected prefix: %s", prefix)
	}

	return nil
}

func (t *Transaction) hasError() bool {
	return t.err != nil
}

func (t Transaction) Error() error {
	return t.err
}
