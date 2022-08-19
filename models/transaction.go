package models

import (
	"encoding/hex"
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

func (t *Transaction) SetSignature(sigType string, signature interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	t.Signature = &Signature{}

	switch sigType {
	case "ed25519_signature", "multi_ed25519_signature", "multi_agent_signature":
		t.Signature.Type = sigType
	default:
		t.err = fmt.Errorf("unexpected sig type %s", sigType)
		return t
	}

	switch sig := signature.(type) {
	case ED25519Signature:
		t.Signature.ED25519Signature = sig
		if err := t.validateED25519(t.Signature.ED25519Signature); err != nil {
			t.err = err
			return t
		}
	case MultiED25519Signature:
		t.Signature.MultiED25519Signature = sig
		if err := t.validateMultiED25519(t.Signature.MultiED25519Signature); err != nil {
			t.err = err
			return t
		}
	case MultiAgentSignature:
		t.Signature.MultiAgentSignature = sig
		if err := t.validateMultiAgent(t.Signature.MultiAgentSignature); err != nil {
			t.err = err
			return t
		}
	default:
		t.err = fmt.Errorf("unexpected signature type %T", sig)
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

func (t Transaction) TxForSimulate() UserTransaction {
	tx := t.UserTransaction
	zeroSig := "0x" + strings.Repeat("00", 64)
	tx.Signature = &Signature{
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

		newSecondarySigners := make([]struct {
			Type string `json:"type"`
			ED25519Signature
			MultiED25519Signature
		}, len(tx.Signature.SecondarySigners))
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

func (t Transaction) validatePub(pub string) error {
	pub = strings.TrimPrefix(pub, "0x")
	pub = strings.TrimPrefix(pub, "0X")
	pubBytes, err := hex.DecodeString(pub)
	if err != nil {
		return err
	}

	if len(pubBytes) != 32 {
		return fmt.Errorf("incorrect pub length %d", len(pubBytes))
	}

	return nil
}

func (t Transaction) validateSig(sig string) error {
	sig = strings.TrimPrefix(sig, "0x")
	sig = strings.TrimPrefix(sig, "0X")
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return err
	}

	if len(sigBytes) != 64 {
		return fmt.Errorf("incorrect sig length %d", len(sigBytes))
	}

	return nil
}

func (t Transaction) validateBitmap(bitmap string) error {
	bitmap = strings.TrimPrefix(bitmap, "0x")
	bitmap = strings.TrimPrefix(bitmap, "0X")
	bitmapBytes, err := hex.DecodeString(bitmap)
	if err != nil {
		return err
	}

	if len(bitmapBytes) != 4 {
		return fmt.Errorf("incorrect bitmap length %d", len(bitmapBytes))
	}

	return nil
}

func (t Transaction) validateSigWithPub(sig, pub string) error {
	pub = strings.TrimPrefix(pub, "0x")
	pub = strings.TrimPrefix(pub, "0X")
	pubBytes, err := hex.DecodeString(pub)
	if err != nil {
		return err
	}

	sig = strings.TrimPrefix(sig, "0x")
	sig = strings.TrimPrefix(sig, "0X")
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubBytes, t.signingMessage, sigBytes) {
		return fmt.Errorf("ed25519.Verify failed")
	}

	return nil
}

func (t Transaction) validateED25519(ed25519Sig ED25519Signature) error {
	if err := t.validatePub(ed25519Sig.PublicKey); err != nil {
		return err
	}

	if err := t.validateSig(ed25519Sig.Signature); err != nil {
		return err
	}

	if err := t.validateSigWithPub(ed25519Sig.Signature, ed25519Sig.PublicKey); err != nil {
		return err
	}

	return nil
}

func (t Transaction) validateMultiED25519(multiED25519Sig MultiED25519Signature) error {
	if len(multiED25519Sig.Signatures) < multiED25519Sig.Threshold {
		return fmt.Errorf("signatures size(%d) must >= threshold(%d)",
			len(multiED25519Sig.Signatures), multiED25519Sig.Threshold)
	}

	for _, sig := range multiED25519Sig.Signatures {
		if err := t.validateSig(sig); err != nil {
			return err
		}
	}

	for _, pub := range multiED25519Sig.PublicKeys {
		if err := t.validatePub(pub); err != nil {
			return err
		}
	}

	if err := t.validateBitmap(multiED25519Sig.Bitmap); err != nil {
		return err
	}

	bitmapBytes, err := hex.DecodeString(multiED25519Sig.Bitmap)
	if err != nil {
		return err
	}

	var sigIndex int
	bitmapBigInt := new(big.Int).SetBytes(bitmapBytes)
	for i := 0; i < 32; i++ {
		if bitmapBigInt.Bit(31-i) == 1 {
			if i > len(multiED25519Sig.PublicKeys)-1 {
				return fmt.Errorf("bitmap %b and public keys length %d not match",
					bitmapBytes, len(multiED25519Sig.PublicKeys))
			}
			if err := t.validateSigWithPub(
				multiED25519Sig.Signatures[sigIndex], multiED25519Sig.PublicKeys[i]); err != nil {
				return err
			}
			sigIndex++
		}
	}

	if sigIndex != len(multiED25519Sig.Signatures) {
		return fmt.Errorf("does not have enough signatures: %d vs %d",
			sigIndex, len(multiED25519Sig.Signatures))
	}
	return nil
}

func (t Transaction) validateMultiAgent(multiAgentSig MultiAgentSignature) error {
	switch multiAgentSig.Sender.Type {
	case "ed25519_signature":
		if err := t.validateED25519(multiAgentSig.Sender.ED25519Signature); err != nil {
			return err
		}
	case "multi_ed25519_signature":
		if err := t.validateMultiED25519(multiAgentSig.Sender.MultiED25519Signature); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected sig type %s", multiAgentSig.Sender.Type)
	}

	if len(multiAgentSig.SecondarySigners) != len(multiAgentSig.SecondarySigners) {
		return fmt.Errorf("incorrect agent signatures size: %d vs %d",
			len(multiAgentSig.SecondarySigners), len(multiAgentSig.SecondarySigners))
	}

	for _, sig := range multiAgentSig.SecondarySigners {
		switch sig.Type {
		case "ed25519_signature":
			if err := t.validateED25519(sig.ED25519Signature); err != nil {
				return err
			}
		case "multi_ed25519_signature":
			if err := t.validateMultiED25519(sig.MultiED25519Signature); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected sig type %s", sig.Type)
		}
	}

	return nil
}

// "0xb5e97db07fa0bd0e5598aa3643a9bc6f6693bddc1a9fec9e674a461eaa00b193"
var RawTransactionSalt = sha3.Sum256([]byte("APTOS::RawTransaction"))

// "0x5efa3c4f02f83a0f4b2d69fc95c607cc02825cc4e7be536ef0992df050d9e67c"
var RawTransactionWithDataSalt = sha3.Sum256([]byte("APTOS::RawTransactionWithData"))

func (t Transaction) GetSigningMessage() ([]byte, error) {
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

		return append(RawTransactionWithDataSalt[:], bcsBytes...), nil
	} else {
		bcsBytes, err := lcs.Marshal(t.RawTransaction)
		if err != nil {
			return nil, err
		}

		return append(RawTransactionSalt[:], bcsBytes...), nil
	}

}

func (t *Transaction) SetSigningMessage(signingMessage string) *Transaction {
	if t.hasError() {
		return t
	}

	signingMessage = strings.TrimPrefix(signingMessage, "0x")
	msg, err := hex.DecodeString(signingMessage)
	if err != nil {
		t.err = err
		return t
	}

	t.signingMessage = msg
	return t
}

func (t *Transaction) hasError() bool {
	return t.err != nil
}

func (t Transaction) Error() error {
	return t.err
}
