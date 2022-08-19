package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/ed25519"

	"github.com/portto/aptos-go-sdk/client"
)

type Transaction struct {
	err error

	// signing message for verification
	SigningMessage string `json:"signing_message"`
	client.Transaction
}

func (t *Transaction) SetSender(sender string) *Transaction {
	if t.hasError() {
		return t
	}

	sender = strings.TrimPrefix(sender, "0x")
	sender = strings.TrimPrefix(sender, "0X")
	senderBytes, err := hex.DecodeString(sender)
	if err != nil {
		t.err = err
		return t
	}

	length := len(senderBytes)
	if length != 32 {
		t.err = fmt.Errorf("unexpected sender length: %d", length)
		return t
	}

	t.Sender = "0x" + hex.EncodeToString(senderBytes)
	return t
}

func (t *Transaction) SetSequenceNumber(seq interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch seq.(type) {
	case uint64:
		t.SequenceNumber = new(big.Int).SetUint64(seq.(uint64)).String()
	case string:
		s, ok := new(big.Int).SetString(seq.(string), 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.SequenceNumber = s.String()
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

	switch maxGasAmount.(type) {
	case uint64:
		t.MaxGasAmount = new(big.Int).SetUint64(maxGasAmount.(uint64)).String()
	case string:
		s, ok := new(big.Int).SetString(maxGasAmount.(string), 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.MaxGasAmount = s.String()
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

	switch gasUnitPrice.(type) {
	case uint64:
		t.GasUnitPrice = new(big.Int).SetUint64(gasUnitPrice.(uint64)).String()
	case string:
		s, ok := new(big.Int).SetString(gasUnitPrice.(string), 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.GasUnitPrice = s.String()
	default:
		t.err = fmt.Errorf("unexpected type: %T", gasUnitPrice)
		return t
	}

	return t
}

func (t *Transaction) SetGasCurrencyCode(gasCurrencyCode string) *Transaction {
	if t.hasError() {
		return t
	}

	t.GasCurrencyCode = gasCurrencyCode
	return t
}

func (t *Transaction) SetExpirationTimestampSecs(expirationTimestampSecs interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch expirationTimestampSecs.(type) {
	case uint64:
		t.ExpirationTimestampSecs = new(big.Int).SetUint64(expirationTimestampSecs.(uint64)).String()
	case string:
		s, ok := new(big.Int).SetString(expirationTimestampSecs.(string), 10)
		if !ok {
			t.err = fmt.Errorf("seq new(big.Int).SetString failed")
			return t
		}

		t.ExpirationTimestampSecs = s.String()
	default:
		t.err = fmt.Errorf("unexpected type: %T", expirationTimestampSecs)
		return t
	}
	return t
}

func (t *Transaction) SetPayload(payloadType string, typeArgs []string, args []interface{}, payload interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	switch payloadType {
	case "entry_function_payload", "script_payload",
		"module_bundle_payload", "write_set_payload":
		t.Payload.Type = payloadType
		t.Payload.TypeArguments = typeArgs
		t.Payload.Arguments = args
	default:
		t.err = fmt.Errorf("unexpected type %s", payloadType)
		return t
	}

	switch payload.(type) {
	case client.ScriptFunctionPayload:
		t.Payload.ScriptFunctionPayload = payload.(client.ScriptFunctionPayload)
	case client.ScriptPayload:
		t.Payload.ScriptPayload = payload.(client.ScriptPayload)
	case client.ModuleBundlePayload:
		t.Payload.ModuleBundlePayload = payload.(client.ModuleBundlePayload)
	case client.WriteSetPayload:
		t.Payload.WriteSetPayload = payload.(client.WriteSetPayload)
	default:
		t.err = fmt.Errorf("unexpected payload type %T", payloadType)
		return t
	}

	return t
}

func (t *Transaction) SetSignature(sigType string, signature interface{}) *Transaction {
	if t.hasError() {
		return t
	}

	t.Signature = &struct {
		Type string `json:"type"`
		client.MultiED25519Signature
		client.ED25519Signature
		client.MultiAgentSignature
	}{}

	switch sigType {
	case "ed25519_signature", "multi_ed25519_signature", "multi_agent_signature":
		t.Signature.Type = sigType
	default:
		t.err = fmt.Errorf("unexpected sig type %s", sigType)
		return t
	}

	switch signature.(type) {
	case client.ED25519Signature:
		t.Signature.ED25519Signature = signature.(client.ED25519Signature)
		if err := t.validateED25519(t.Signature.ED25519Signature); err != nil {
			t.err = err
			return t
		}
	case client.MultiED25519Signature:
		t.Signature.MultiED25519Signature = signature.(client.MultiED25519Signature)
		if err := t.validateMultiED25519(t.Signature.MultiED25519Signature); err != nil {
			t.err = err
			return t
		}
	case client.MultiAgentSignature:
		t.Signature.MultiAgentSignature = signature.(client.MultiAgentSignature)
		if err := t.validateMultiAgent(t.Signature.MultiAgentSignature); err != nil {
			t.err = err
			return t
		}
	default:
		t.err = fmt.Errorf("unexpected signature type %T", signature)
		return t
	}

	return t
}

func (t *Transaction) SetSecondarySigners(secondarySigners []string) *Transaction {
	if t.hasError() {
		return t
	}

	t.SecondarySigners = secondarySigners
	return t
}

func (t Transaction) TxForSimulate() client.Transaction {
	tx := t.Transaction
	zeroSig := "0x" + strings.Repeat("00", 64)
	tx.Signature = &struct {
		Type string `json:"type"`
		client.MultiED25519Signature
		client.ED25519Signature
		client.MultiAgentSignature
	}{
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
			client.ED25519Signature
			client.MultiED25519Signature
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

	sigMsg := strings.TrimPrefix(t.SigningMessage, "0x")
	sigMsg = strings.TrimPrefix(sigMsg, "0X")
	sigMsgBytes, err := hex.DecodeString(sigMsg)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubBytes, sigMsgBytes, sigBytes) {
		return fmt.Errorf("ed25519.Verify failed")
	}

	return nil
}

func (t Transaction) validateED25519(ed25519Sig client.ED25519Signature) error {
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

func (t Transaction) validateMultiED25519(multiED25519Sig client.MultiED25519Signature) error {
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

func (t Transaction) validateMultiAgent(multiAgentSig client.MultiAgentSignature) error {
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

func (t *Transaction) SetSigningMessage(signingMessage string) *Transaction {
	if t.hasError() {
		return t
	}

	signingMessage = strings.TrimPrefix(signingMessage, "0x")
	signingMessage = strings.TrimPrefix(signingMessage, "0X")
	_, err := hex.DecodeString(signingMessage)
	if err != nil {
		t.err = err
		return t
	}

	t.SigningMessage = signingMessage
	return t
}

func (t *Transaction) hasError() bool {
	return t.err != nil
}

func (t Transaction) Error() error {
	return t.err
}
