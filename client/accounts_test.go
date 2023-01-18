package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	mockAddr    = "0xa4a793df771b6f48af4d9dbbe35ef2137a5ff0d7217ac5cc14544e4f30522a78"
	modulesResp = `[{"bytecode":"test_bytecode","abi":{"address":"0xa4a793df771b6f48af4d9dbbe35ef2137a5ff0d7217ac5cc14544e4f30522a78","name":"test","friends":[],"exposed_functions":[{"name":"test","visibility":"public","is_entry":true,"generic_type_params":[],"params":["vector<u8>"],"return":["vector<u8>"]}],"structs":[{"name":"TestEvent","is_native":false,"abilities":["drop","store"],"generic_type_params":[{"constraints":[]}],"fields":[{"name":"amount","type":"u64"}]}]}}]`
	errResp     = `{"message":"Account not found by Address(0xa3a793df771b6f48af4d9dbbe35ef2137a5ff0d7217ac5cc14544e4f30522a7) and Ledger version(21067593)","error_code":"account_not_found","vm_error_code":null}`
)

var (
	ctx = context.Background()
)

func TestGetModules(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, err := w.Write([]byte(modulesResp))
			assert.NoError(t, err)
		}))

		c := NewAptosClient(srv.URL)
		resp, err := c.GetAccountModules(ctx, mockAddr)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp))
		assert.Equal(t, AccountModule{
			Bytecode: "test_bytecode",
			ABI: ABI{
				Address: mockAddr,
				Name:    "test",
				Friends: []string{},
				ExposedFunctions: []ExposedFunction{
					{
						Name:              "test",
						Visibility:        "public",
						IsEntry:           true,
						GenericTypeParams: []GenericTypeParam{},
						Params:            []string{"vector<u8>"},
						Return:            []string{"vector<u8>"},
					},
				},
				Structs: []Struct{
					{
						Name:      "TestEvent",
						IsNative:  false,
						Abilities: []string{"drop", "store"},
						GenericTypeParams: []GenericTypeParam{
							{
								Constraints: []string{},
							},
						},
						Fields: []Field{
							{
								Name: "amount",
								Type: "u64",
							},
						},
					},
				},
			},
		}, resp[0])
	})

	t.Run("AccountNotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(errResp))
		}))

		c := NewAptosClient(srv.URL)
		_, err := c.GetAccountModules(ctx, mockAddr)
		var e *Error
		assert.Equal(t, true, errors.As(err, &e))
		assert.Equal(t, true, e.IsErrorCode(ErrAccountNotFound))
	})
}
