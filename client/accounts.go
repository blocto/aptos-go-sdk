package client

import (
	"context"
	"fmt"
	"net/http"
)

type Accounts interface {
	GetAccount(ctx context.Context, address string, opts ...interface{}) (*AccountInfo, error)
	GetAccountResources(ctx context.Context, address string, opts ...interface{}) ([]AccountResource, error)
	GetResourceByAccountAddressAndResourceType(ctx context.Context, address, resourceType string, opts ...interface{}) (*AccountResource, error)
	GetAccountModules(ctx context.Context, address string, opts ...interface{}) ([]AccountModule, error)
	GetModuleByModuleID(ctx context.Context, address, moduleID string, opts ...interface{}) (*AccountModule, error)
}

type AccountsImpl struct {
	Base
}

type AccountInfo struct {
	SequenceNumber    string `json:"sequence_number"`
	AuthenticationKey string `json:"authentication_key"`
}

func (impl AccountsImpl) GetAccount(ctx context.Context, address string, opts ...interface{}) (*AccountInfo, error) {
	var rspJSON AccountInfo
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s", address),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type AccountResource struct {
	Type string
	Data interface{}
}

func (impl AccountsImpl) GetAccountResources(ctx context.Context, address string, opts ...interface{}) ([]AccountResource, error) {
	var rspJSON []AccountResource
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/resources", address),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetResourceByAccountAddressAndResourceType(ctx context.Context, address, resourceType string, opts ...interface{}) (*AccountResource, error) {
	var rspJSON AccountResource
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/resource/%s", address, resourceType),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type AccountModule struct {
	Bytecode string      `json:"bytecode"`
	ABI      interface{} `json:"abi"`
}

func (impl AccountsImpl) GetAccountModules(ctx context.Context, address string, opts ...interface{}) ([]AccountModule, error) {
	var rspJSON []AccountModule
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/modules", address),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetModuleByModuleID(ctx context.Context, address, moduleID string, opts ...interface{}) (*AccountModule, error) {
	var rspJSON AccountModule
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/module/%s", address, moduleID),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
