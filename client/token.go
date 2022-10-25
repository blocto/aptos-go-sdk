package client

import (
	"context"
	"encoding/json"
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

	GetCollectionData(ctx context.Context, creator models.AccountAddress, collectionName string) (*models.CollectionData, error)
	GetTokenData(ctx context.Context, creator models.AccountAddress, collectionName, tokenName string) (*models.TokenData, error)
	GetToken(ctx context.Context, owner models.AccountAddress, tokenID models.TokenID) (*models.Token, error)
	ListAccountTokens(ctx context.Context, owner models.AccountAddress) ([]models.Token, error)
}

// NewTokenClient creates TokenClient to do things with aptos token.
func NewTokenClient(client AptosClient) (TokenClient, error) {
	ledgerInfo, err := client.LedgerInformation(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get ledger info error: %w", err)
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
	MutateConfig models.CollectionMutabilityConfig
}

func (impl *TokenClientImpl) CreateCollection(ctx context.Context, creator models.SingleSigner, req CreateCollectionRequest) (string, error) {
	tx := models.Transaction{}

	addr := creator.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %w", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %w", err)
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
		return "", fmt.Errorf("submit tx error: %w", err)
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
	MutateConfig             models.TokenMutabilityConfig
	PropertyKeys             []string
	PropertyValues           []string
	PropertyTypes            []string
}

func (impl *TokenClientImpl) CreateToken(ctx context.Context, creator models.SingleSigner, req CreateTokenRequest) (string, error) {
	tx := models.Transaction{}

	addr := creator.AccountAddress.PrefixZeroTrimmedHex()

	accountInfo, err := impl.client.GetAccount(ctx, addr)
	if err != nil {
		return "", fmt.Errorf("get account info error: %w", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %w", err)
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
		return "", fmt.Errorf("submit tx error: %w", err)
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
		return "", fmt.Errorf("get account info error: %w", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %w", err)
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
		return "", fmt.Errorf("submit tx error: %w", err)
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
		return "", fmt.Errorf("get account info error: %w", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %w", err)
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
		return "", fmt.Errorf("submit tx error: %w", err)
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
		return "", fmt.Errorf("get account info error: %w", err)
	}

	gasPrice, err := impl.client.EstimateGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("get estimate gas price error: %w", err)
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

func (impl *TokenClientImpl) GetCollectionData(ctx context.Context, creator models.AccountAddress, collectionName string) (*models.CollectionData, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections",
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %w", err)
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

	var data models.CollectionData
	if err := impl.client.GetTableItemByHandleAndKey(ctx, collectionsHandle, req, &data); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %w", err)
	}

	return &data, nil
}

func (impl *TokenClientImpl) GetTokenData(ctx context.Context, creator models.AccountAddress, collectionName, tokenName string) (*models.TokenData, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections",
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %w", err)
	}

	if resource.Data.CollectionsResource == nil {
		return nil, errors.New("nil CollectionsResource")
	}

	tokensHandle := resource.Data.TokenData.Handle

	req := TableItemReq{
		KeyType:   "0x3::token::TokenDataId",
		ValueType: "0x3::token::TokenData",
		Key: models.TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: collectionName,
			Name:       tokenName,
		},
	}

	var data models.TokenData
	if err := impl.client.GetTableItemByHandleAndKey(ctx, tokensHandle, req, &data); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %w", err)
	}

	return &data, nil
}

const tokenStoreType = "0x3::token::TokenStore"

func (impl *TokenClientImpl) GetToken(ctx context.Context, owner models.AccountAddress, tokenID models.TokenID) (*models.Token, error) {
	resource, err := impl.client.GetResourceByAccountAddressAndResourceType(
		ctx, owner.PrefixZeroTrimmedHex(), tokenStoreType,
	)
	if err != nil {
		return nil, fmt.Errorf("client.GetResourceByAccountAddressAndResourceType error: %w", err)
	}

	if resource.Data.TokenStoreResource == nil {
		return nil, errors.New("nil TokenStoreResource")
	}

	tokenStoreHandle := resource.Data.Tokens.Handle

	req := TableItemReq{
		KeyType:   "0x3::token::TokenId",
		ValueType: "0x3::token::Token",
		Key:       tokenID,
	}

	var token models.Token
	if err := impl.client.GetTableItemByHandleAndKey(ctx, tokenStoreHandle, req, &token); err != nil {
		return nil, fmt.Errorf("client.GetTableItemByHandleAndKey error: %w", err)
	}

	return &token, nil
}

const depositEventsField = "deposit_events"

// ListAccountTokens gets all tokens of the owner by fetching all token deposit events.
// It returns an error if the client fails to get the deposit events or the owner has
// tokens but the client fails to get one of them.
func (impl *TokenClientImpl) ListAccountTokens(ctx context.Context, owner models.AccountAddress) ([]models.Token, error) {
	var tokens []models.Token

	tokenIDs := make(map[models.TokenID]bool)

	var start uint64 = 0
	var limit uint64 = 100

	for {
		events, err := impl.client.GetEventsByEventHandle(
			ctx, owner.PrefixZeroTrimmedHex(),
			tokenStoreType, depositEventsField, start, limit,
		)
		if err != nil {
			return nil, fmt.Errorf("client.GetEventsByEventHandle error: %w", err)
		}

		if len(events) == 0 {
			break
		}

		for _, event := range events {
			start = uint64(event.SequenceNumber) + 1

			var data models.TokenDepositEvent
			b, err := json.Marshal(event.Data)
			if err != nil {
				return nil, fmt.Errorf("json.Marshal event data %+v error: %w", event.Data, err)
			}
			if err := json.Unmarshal(b, &data); err != nil {
				return nil, fmt.Errorf("json.Unmarshal into TokenDepositEvent error: %w", err)
			}

			if !tokenIDs[data.ID] {
				token, err := impl.GetToken(ctx, owner, data.ID)
				if err != nil {
					if err, ok := errors.Unwrap(err).(Error); ok {
						if err.IsTableItemNotFound() {
							continue
						}
					}
					return nil, fmt.Errorf("GetToken error: %v", err)
				}
				tokens = append(tokens, *token)
				tokenIDs[data.ID] = true
			}
		}
	}

	return tokens, nil
}
