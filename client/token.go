package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/portto/aptos-go-sdk/models"
)

type TokenClient interface {
	CreateCollection(ctx context.Context, creator models.SingleSigner, req CreateCollectionRequest) (string, error)
	CreateToken(ctx context.Context, creator models.SingleSigner, req CreateTokenRequest) (string, error)
	MintToken(ctx context.Context, minter models.SingleSigner, req MintTokenRequest) (string, error)
	OfferToken(ctx context.Context, sender models.SingleSigner, req OfferTokenRequest) (string, error)
	ClaimToken(ctx context.Context, receiver models.SingleSigner, req ClaimTokenRequest) (string, error)

	GetCollectionData(ctx context.Context, creator models.AccountAddress, collectionName string) (*CollectionData, error)
	GetTokenData(ctx context.Context, creator models.AccountAddress, collectionName, tokenName string) (*TokenData, error)
	GetToken(ctx context.Context, owner models.AccountAddress, tokenID TokenID) (*Token, error)
}

// NewTokenClient creates TokenClient to do things with aptos token.
func NewTokenClient(client AptosClient) (TokenClient, error) {
	ledgerInfo, err := client.LedgerInformation(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get ledger info error: %v", err)
	}

	return &TokenClientImpl{
		client:  client,
		chainID: ledgerInfo.ChainID,
	}, nil
}

type TokenClientImpl struct {
	client  AptosClient
	chainID uint8
}

var DefaultMaxGasAmount uint64 = 5000

var TokenModule models.Module
var TokenTransferModule models.Module

func init() {
	moduleAddr, _ := models.HexToAccountAddress("0x3")
	TokenModule = models.Module{
		Address: moduleAddr,
		Name:    "token",
	}
	TokenTransferModule = models.Module{
		Address: moduleAddr,
		Name:    "token_transfers",
	}
}

type CreateCollectionRequest struct {
	Name         string
	Description  string
	URI          string
	Maximum      uint64
	MutateConfig CollectionMutabilityConfig
}

type CollectionMutabilityConfig struct {
	Description bool
	URI         bool
	Maximum     bool
}

func (impl *TokenClientImpl) CreateCollection(ctx context.Context, creator models.SingleSigner, req CreateCollectionRequest) (string, error) {
	tx := models.Transaction{}

	addr := creator.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %v", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %v", err)
	}

	err = tx.SetChainID(impl.chainID).
		SetSender(addr).
		SetPayload(models.EntryFunctionPayload{
			Module:   TokenModule,
			Function: "create_collection_script",
			Arguments: []interface{}{req.Name, req.Description, req.URI, req.Maximum,
				[]bool{req.MutateConfig.Description, req.MutateConfig.Maximum, req.MutateConfig.URI}},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(30 * time.Second).Unix())).
		SetGasUnitPrice(gasPrice).
		SetMaxGasAmount(DefaultMaxGasAmount).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()

	if err != nil {
		return "", fmt.Errorf("build tx error: %v", err)
	}

	if err := creator.Sign(&tx).Error(); err != nil {
		return "", fmt.Errorf("sign tx error: %v", err)
	}

	txResp, err := impl.client.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		return "", fmt.Errorf("submit tx error: %v", err)
	}

	return txResp.Hash, nil
}

type CreateTokenRequest struct {
	Collection               string
	Name                     string
	Description              string
	Supply                   uint64
	Maximum                  uint64
	URI                      string
	RoyaltyPayeeAddress      models.AccountAddress
	RoyaltyPointsDenominator uint64
	RoyaltyPointsNumerator   uint64
	MutateConfig             TokenMutabilityConfig
	PropertyKeys             []string
	PropertyValues           []string
	PropertyTypes            []string
}

type TokenMutabilityConfig struct {
	Maximum     bool
	URI         bool
	Description bool
	Royalty     bool
	Properties  bool
}

func (impl *TokenClientImpl) CreateToken(ctx context.Context, creator models.SingleSigner, req CreateTokenRequest) (string, error) {
	tx := models.Transaction{}

	addr := creator.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %v", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %v", err)
	}

	err = tx.SetChainID(impl.chainID).
		SetSender(addr).
		SetPayload(models.EntryFunctionPayload{
			Module:   TokenModule,
			Function: "create_token_script",
			Arguments: []interface{}{
				req.Collection, req.Name, req.Description, req.Supply, req.Maximum, req.URI,
				req.RoyaltyPayeeAddress, req.RoyaltyPointsDenominator, req.RoyaltyPointsNumerator,
				[]bool{req.MutateConfig.Maximum, req.MutateConfig.URI, req.MutateConfig.Description,
					req.MutateConfig.Royalty, req.MutateConfig.Properties},
				req.PropertyKeys, req.PropertyValues, req.PropertyTypes,
			},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(30 * time.Second).Unix())).
		SetGasUnitPrice(gasPrice).
		SetMaxGasAmount(DefaultMaxGasAmount).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()

	if err != nil {
		return "", fmt.Errorf("build tx error: %v", err)
	}

	if err := creator.Sign(&tx).Error(); err != nil {
		return "", fmt.Errorf("sign tx error: %v", err)
	}

	txResp, err := impl.client.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		return "", fmt.Errorf("submit tx error: %v", err)
	}

	return txResp.Hash, nil
}

type MintTokenRequest struct {
	Creator    models.AccountAddress
	Collection string
	TokenName  string
	Amount     uint64
}

func (impl *TokenClientImpl) MintToken(ctx context.Context, minter models.SingleSigner, req MintTokenRequest) (string, error) {
	tx := models.Transaction{}

	addr := minter.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %v", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %v", err)
	}

	err = tx.SetChainID(impl.chainID).
		SetSender(addr).
		SetPayload(models.EntryFunctionPayload{
			Module:    TokenModule,
			Function:  "mint_script",
			Arguments: []interface{}{req.Creator, req.Collection, req.TokenName, req.Amount},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(30 * time.Second).Unix())).
		SetGasUnitPrice(gasPrice).
		SetMaxGasAmount(DefaultMaxGasAmount).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()

	if err != nil {
		return "", fmt.Errorf("build tx error: %v", err)
	}

	if err := minter.Sign(&tx).Error(); err != nil {
		return "", fmt.Errorf("sign tx error: %v", err)
	}

	txResp, err := impl.client.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		return "", fmt.Errorf("submit tx error: %v", err)
	}

	return txResp.Hash, nil
}

type OfferTokenRequest struct {
	Receiver        models.AccountAddress
	Creator         models.AccountAddress
	Collection      string
	TokenName       string
	PropertyVersion uint64
	Amount          uint64
}

func (impl *TokenClientImpl) OfferToken(ctx context.Context, sender models.SingleSigner, req OfferTokenRequest) (string, error) {
	tx := models.Transaction{}

	addr := sender.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %v", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %v", err)
	}

	err = tx.SetChainID(impl.chainID).
		SetSender(addr).
		SetPayload(models.EntryFunctionPayload{
			Module:   TokenTransferModule,
			Function: "offer_script",
			Arguments: []interface{}{
				req.Receiver,
				req.Creator,
				req.Collection,
				req.TokenName,
				req.PropertyVersion,
				req.Amount,
			},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(30 * time.Second).Unix())).
		SetGasUnitPrice(gasPrice).
		SetMaxGasAmount(DefaultMaxGasAmount).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()

	if err != nil {
		return "", fmt.Errorf("build tx error: %v", err)
	}

	if err := sender.Sign(&tx).Error(); err != nil {
		return "", fmt.Errorf("sign tx error: %v", err)
	}

	txResp, err := impl.client.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		return "", fmt.Errorf("submit tx error: %v", err)
	}

	return txResp.Hash, nil
}

type ClaimTokenRequest struct {
	Sender          models.AccountAddress
	Creator         models.AccountAddress
	Collection      string
	TokenName       string
	PropertyVersion uint64
}

func (impl *TokenClientImpl) ClaimToken(ctx context.Context, receiver models.SingleSigner, req ClaimTokenRequest) (string, error) {
	tx := models.Transaction{}

	addr := receiver.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %v", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %v", err)
	}

	err = tx.SetChainID(impl.chainID).
		SetSender(addr).
		SetPayload(models.EntryFunctionPayload{
			Module:   TokenTransferModule,
			Function: "claim_script",
			Arguments: []interface{}{
				req.Sender,
				req.Creator,
				req.Collection,
				req.TokenName,
				req.PropertyVersion,
			},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(30 * time.Second).Unix())).
		SetGasUnitPrice(gasPrice).
		SetMaxGasAmount(DefaultMaxGasAmount).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()

	if err != nil {
		return "", fmt.Errorf("build tx error: %v", err)
	}

	if err := receiver.Sign(&tx).Error(); err != nil {
		return "", fmt.Errorf("sign tx error: %v", err)
	}

	txResp, err := impl.client.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		return "", fmt.Errorf("submit tx error: %v", err)
	}

	return txResp.Hash, nil
}

type CollectionData struct {
	Name         string
	Description  string
	URI          string
	Maximum      string                     // uint64
	Supply       string                     // uint64
	MutateConfig CollectionMutabilityConfig `json:"mutability_config"`
}

func (impl *TokenClientImpl) GetCollectionData(ctx context.Context, creator models.AccountAddress, collectionName string) (*CollectionData, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections",
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %v", err)
	}

	if resource.Data.CollectionsResource == nil {
		return nil, errors.New("nil CollectionsResource")
	}

	collectionsHandle := resource.Data.CollectionData.Handle

	req := TableItemReq{
		KeyType:   "0x1::string::String",
		ValueType: "0x3::token::CollectionData",
		Key:       collectionName,
	}

	var data CollectionData
	if err := impl.client.GetTableItemByHandleAndKey(ctx, collectionsHandle, req, &data); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %v", err)
	}

	return &data, nil
}

type TokenDataID struct {
	Creator    string `json:"creator"`
	Collection string `json:"collection"`
	Name       string `json:"name"`
}

type TokenData struct {
	Collection   string                `json:"collection"`
	Description  string                `json:"description"`
	Name         string                `json:"name"`
	Maximum      string                `json:"maximum"` // uint64
	Supply       string                `json:"supply"`  // uint64
	URI          string                `json:"uri"`
	MutateConfig TokenMutabilityConfig `json:"mutability_config"`
}

func (impl *TokenClientImpl) GetTokenData(ctx context.Context, creator models.AccountAddress, collectionName, tokenName string) (*TokenData, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections",
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %v", err)
	}

	if resource.Data.CollectionsResource == nil {
		return nil, errors.New("nil CollectionsResource")
	}

	tokensHandle := resource.Data.TokenData.Handle

	req := TableItemReq{
		KeyType:   "0x3::token::TokenDataId",
		ValueType: "0x3::token::TokenData",
		Key: TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: collectionName,
			Name:       tokenName,
		},
	}

	var data TokenData
	if err := impl.client.GetTableItemByHandleAndKey(ctx, tokensHandle, req, &data); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %v", err)
	}

	return &data, nil
}

type TokenID struct {
	TokenDataID     `json:"token_data_id"`
	PropertyVersion string `json:"property_version"`
}

type Token struct {
	ID              TokenID                `json:"id"`
	Amount          string                 `json:"amount"`
	TokenProperties map[string]interface{} `json:"token_properties"`
}

func (impl *TokenClientImpl) GetToken(ctx context.Context, owner models.AccountAddress, tokenID TokenID) (*Token, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, owner.PrefixZeroTrimmedHex(), "0x3::token::TokenStore",
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %v", err)
	}

	if resource.Data.TokenStoreResource == nil {
		return nil, errors.New("nil TokenStoreResource")
	}

	if tokenID.PropertyVersion == "" {
		tokenID.PropertyVersion = "0"
	}

	tokenStoreHandle := resource.Data.Tokens.Handle

	req := TableItemReq{
		KeyType:   "0x3::token::TokenId",
		ValueType: "0x3::token::Token",
		Key:       tokenID,
	}

	var token Token
	if err := impl.client.GetTableItemByHandleAndKey(ctx, tokenStoreHandle, req, &token); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %v", err)
	}

	return &token, nil
}
