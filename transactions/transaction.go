package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"aptos-go-sdk/client"
)

type Transaction struct {
	err error

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
	if length != 16 && length != 32 {
		t.err = fmt.Errorf("unexpected sender length: %d", length)
		return t
	}

	t.Sender = "0x" + hex.EncodeToString(senderBytes[:16])
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
	case "script_function_payload", "script_payload",
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

func (t *Transaction) hasError() bool {
	return t.err != nil
}
