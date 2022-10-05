package models

import (
	"github.com/the729/lcs"
)

type TransactionPayload interface{}

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

type ModuleBundlePayload struct {
	Modules []struct {
		Code []byte
	}
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

type JSONPayload struct {
	Type          string        `json:"type"`
	TypeArguments []string      `json:"type_arguments"`
	Arguments     []interface{} `json:"arguments"`

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
	Version string `json:"version"`
	GUID    struct {
		CreationNumber string `json:"creation_number"`
		AccountAddress string `json:"account_address"`
	} `json:"guid"`
	SequenceNumber string                 `json:"sequence_number"`
	Type           string                 `json:"type"`
	Data           map[string]interface{} `json:"data"`
}
