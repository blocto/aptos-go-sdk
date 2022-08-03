package client

import (
	"fmt"
	"net/http"
)

type Accounts interface {
	GetAccount(address string) (*AccountInfo, error)
	GetAccountResources(address string) ([]AccountResource, error)
	GetResourceByAccountAddressAndResourceType(address, resourceType string) (*AccountResource, error)
	GetAccountModules(address string) ([]AccountModule, error)
	GetModuleByModuleID(address, moduleID string) (*AccountModule, error)
}

type AccountsImpl struct {
	Base
}

type AccountInfo struct {
	SequenceNumber    string `json:"sequence_number"`
	AuthenticationKey string `json:"authentication_key"`
}

func (impl AccountsImpl) GetAccount(address string) (*AccountInfo, error) {
	var rspJSON AccountInfo
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s", address),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type AccountResource struct {
	Type string
	Data interface{}
}

func (impl AccountsImpl) GetAccountResources(address string) ([]AccountResource, error) {
	var rspJSON []AccountResource
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/resources", address),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetResourceByAccountAddressAndResourceType(address, resourceType string) (*AccountResource, error) {
	var rspJSON AccountResource
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/resource/%s", address, resourceType),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type AccountModule struct {
	Bytecode string      `json:"bytecode"`
	ABI      interface{} `json:"abi"`
}

func (impl AccountsImpl) GetAccountModules(address string) ([]AccountModule, error) {
	var rspJSON []AccountModule
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/modules", address),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetModuleByModuleID(address, moduleID string) (*AccountModule, error) {
	var rspJSON AccountModule
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/module/%s", address, moduleID),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
