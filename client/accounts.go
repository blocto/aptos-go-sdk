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
