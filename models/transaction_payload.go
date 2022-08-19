package models

import (
	"encoding/hex"
	"fmt"

	"github.com/the729/lcs"
)

type TransactionPayload interface {
	ToJSON() JSONPayload
}

var _ = lcs.RegisterEnum(
	(*TransactionPayload)(nil),
	ScriptPayload{},
	ModuleBundlePayload{},
	EntryFunctionPayload{},
)

type ScriptPayload struct {
	Code          []byte
	TypeArguments []TypeTag
	Arguments     []TransactionArgument
}

func (p ScriptPayload) ToJSON() JSONPayload {
	json := JSONPayload{
		Type: "script_payload",
		Code: Code{
			Bytecode: "0x" + hex.EncodeToString(p.Code),
		},
	}

	json.TypeArguments = make([]string, len(p.TypeArguments))
	for i, typeArg := range p.TypeArguments {
		json.TypeArguments[i] = typeArg.ToString()
	}

	json.Arguments = make([]string, len(p.Arguments))
	for i, arg := range p.Arguments {
		json.Arguments[i] = arg.ToString()
	}

	return json
}

type ModuleBundlePayload struct {
	Modules []struct {
		Code []byte
	}
}

func (p ModuleBundlePayload) ToJSON() JSONPayload {
	// TODO: implement me
	return JSONPayload{Type: "module_bundle_payload"}
}

type EntryFunctionPayload struct {
	Module
	Function      string
	TypeArguments []TypeTag
	ArgumentsBCS  [][]byte
	Arguments     []interface{} `lcs:"-"`
}

type Module struct {
	Address AccountAddress
	Name    string
}

func (p EntryFunctionPayload) ToJSON() JSONPayload {
	json := JSONPayload{
		Type:     "entry_function_payload",
		Function: fmt.Sprintf("%s::%s::%s", p.Address.PrefixZeroTrimmedHex(), p.Name, p.Function),
	}

	json.TypeArguments = make([]string, len(p.TypeArguments))
	for i, typeArg := range p.TypeArguments {
		json.TypeArguments[i] = typeArg.ToString()
	}

	json.Arguments = make([]string, len(p.Arguments))
	for i, arg := range p.Arguments {
		switch arg := arg.(type) {
		case AccountAddress:
			json.Arguments[i] = hex.EncodeToString(arg[:])
		case [32]byte:
			json.Arguments[i] = hex.EncodeToString(arg[:])
		case []byte:
			json.Arguments[i] = hex.EncodeToString(arg)
		case string:
			json.Arguments[i] = hex.EncodeToString([]byte(arg))
		default:
			json.Arguments[i] = fmt.Sprintf("%v", arg)
		}
	}

	return json
}

type JSONPayload struct {
	Type          string   `json:"type"`
	TypeArguments []string `json:"type_arguments"`
	Arguments     []string `json:"arguments"`

	// ScriptPayload
	Code Code `json:"code,omitempty"`
	// ModuleBundlePayload
	Modules []Code `json:"modules,omitempty"`
	// EntryFunctionPayload
	Function string `json:"function,omitempty"`
}

type Code struct {
	Bytecode string      `json:"bytecode"`
	ABI      interface{} `json:"abi,omitempty"`
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

type Event struct {
	Key            string                 `json:"key"`
	SequenceNumber string                 `json:"sequence_number"`
	Type           string                 `json:"type"`
	Data           map[string]interface{} `json:"data"`
}
