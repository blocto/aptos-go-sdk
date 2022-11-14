package client

import (
	"context"
	"fmt"
	"net/http"
)

type Accounts interface {
	GetAccount(ctx context.Context, address string, opts ...interface{}) (*AccountInfo, error)
	GetAccountResources(ctx context.Context, address string, version int64, opts ...interface{}) ([]AccountResource, error)
	GetResourceByAccountAddressAndResourceType(ctx context.Context, address, resourceType string, version int64, opts ...interface{}) (*AccountResource, error)
	GetAccountModules(ctx context.Context, address string, opts ...interface{}) ([]AccountModule, error)
	GetModuleByModuleID(ctx context.Context, address, moduleID string, opts ...interface{}) (*AccountModule, error)

	GetResourceWithCustomType(ctx context.Context, address, resourceType string, resp interface{}, opts ...interface{}) error
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
	Data struct {
		*CoinStoreResource
		*CollectionsResource
		*TokenStoreResource
	}
}

type Table struct {
	Handle string `json:"handle"`
}

type EventHandle struct {
	Counter string
	GUID    struct {
		ID struct {
			Addr        string `json:"addr"`
			CreationNum string `json:"creation_num"`
		} `json:"id"`
	} `json:"guid"`
}

type CoinStoreResource struct {
	Coin struct {
		Value string `json:"value"`
	} `json:"coin"`
	Frozen         bool        `json:"frozen"`
	DepositEvents  EventHandle `json:"deposit_events"`
	WithdrawEvents EventHandle `json:"withdraw_events"`
}

type CollectionsResource struct {
	CollectionData         Table       `json:"collection_data"`
	TokenData              Table       `json:"token_data"`
	CreateCollectionEvents EventHandle `json:"create_collection_events"`
	CreateTokenDataEvents  EventHandle `json:"create_token_data_events"`
	MintTokenEvents        EventHandle `json:"mint_token_events"`
}

type TokenStoreResource struct {
	DirectTransfer            bool        `json:"direct_transfer"`
	Tokens                    Table       `json:"tokens"`
	DepositEvents             EventHandle `json:"deposit_events"`
	WithdrawEvents            EventHandle `json:"withdraw_events"`
	BurnEvents                EventHandle `json:"burn_events"`
	MutateTokenPropertyEvents EventHandle `json:"mutate_token_property_events"`
}

func (impl AccountsImpl) GetAccountResources(ctx context.Context, address string, version int64, opts ...interface{}) ([]AccountResource, error) {
	var rspJSON []AccountResource
	var url string
	if version < 0 {
		url = fmt.Sprintf("/v1/accounts/%s/resources", address)
	} else {
		url = fmt.Sprintf("/v1/accounts/%s/resources?ledger_version=%d", address, version)
	}
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+url, nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetResourceByAccountAddressAndResourceType(ctx context.Context, address, resourceType string, version int64, opts ...interface{}) (*AccountResource, error) {
	var rspJSON AccountResource
	var url string
	if version < 0 {
		url = fmt.Sprintf("/v1/accounts/%s/resource/%s", address, resourceType)
	} else {
		url = fmt.Sprintf("/v1/accounts/%s/resource/%s?ledger_version=%d", address, resourceType, version)
	}
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+url, nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type AccountModule struct {
	Bytecode string `json:"bytecode"`
	ABI      ABI    `json:"abi"`
}

type ABI struct {
	Address          string            `json:"address"`
	Name             string            `json:"name"`
	Friends          []string          `json:"friends"`
	ExposedFunctions []ExposedFunction `json:"exposed_functions"`
	Structs          []Struct          `json:"structs"`
}

type ExposedFunction struct {
	Name              string             `json:"name"`
	Visibility        string             `json:"visibility"`
	IsEntry           bool               `json:"is_entry"`
	GenericTypeParams []GenericTypeParam `json:"generic_type_params"`
	Params            []string           `json:"params"`
	Return            []string           `json:"return"`
}

type GenericTypeParam struct {
	Constraints []string `json:"constraints"`
}

type Struct struct {
	Name              string             `json:"name"`
	IsNative          bool               `json:"is_native"`
	Abilities         []string           `json:"abilities"`
	GenericTypeParams []GenericTypeParam `json:"generic_type_params"`
	Fields            []Field            `json:"fields"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (impl AccountsImpl) GetAccountModules(ctx context.Context, address string, opts ...interface{}) ([]AccountModule, error) {
	var rspJSON []AccountModule
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/modules", address),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, fmt.Errorf("client.GetAccountModules error: %w", err)
	}

	return rspJSON, nil
}

func (impl AccountsImpl) GetModuleByModuleID(ctx context.Context, address, moduleID string, opts ...interface{}) (*AccountModule, error) {
	var rspJSON AccountModule
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/module/%s", address, moduleID),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, fmt.Errorf("client.GetModuleByModuleID error: %w", err)
	}

	return &rspJSON, nil
}

func (impl AccountsImpl) GetResourceWithCustomType(ctx context.Context, address, resourceType string, resp interface{}, opts ...interface{}) error {
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/resource/%s", address, resourceType),
		nil, resp, nil, requestOptions(opts...))
	if err != nil {
		return err
	}

	return nil
}
