package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"

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
	// ListAccountTokensV2 uses graphql api to get tokens with version v1 and v2.
	ListAccountTokensV2(ctx context.Context, owners ...models.AccountAddress) ([]models.TokenV2, error)
}

// NewTokenClient creates TokenClient to do things with aptos token.
func NewTokenClient(client AptosClient, graphqlEndpoint string) (TokenClient, error) {
	ledgerInfo, err := client.LedgerInformation(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get ledger info error: %w", err)
	}

	return &TokenClientImpl{
		client:  client,
		graphql: graphql.NewClient(graphqlEndpoint, nil),
		chainID: ledgerInfo.ChainID,
	}, nil
}

type TokenClientImpl struct {
	client  AptosClient
	graphql *graphql.Client
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
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections", -1,
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
		ctx, creator.PrefixZeroTrimmedHex(), "0x3::token::Collections", -1,
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
		ctx, owner.PrefixZeroTrimmedHex(), tokenStoreType, -1,
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

// ListAccountTokens gets aptos tokens of an account by indexer graphql api. Returns a list of tokens.
func (impl *TokenClientImpl) ListAccountTokens(ctx context.Context, owner models.AccountAddress) ([]models.Token, error) {
	var tokens []models.Token

	const batchSize = 100

	query := `
	query CurrentTokens($owner_address: String, $offset: Int, $limit: Int) {
		current_token_ownerships(
			where: {owner_address: {_eq: $owner_address}, amount: {_gt: "0"}, table_type: {_eq: "0x3::token::TokenStore"}}
			order_by: {last_transaction_version: asc}
			offset: $offset
			limit: $limit
			) {
				creator_address
				collection_name
				name
				property_version
				amount
				token_properties
				token_data_id_hash
			}
		}
	`
	variables := map[string]interface{}{
		"owner_address": graphql.String(owner.ToHex()),
		"limit":         batchSize,
	}

	for offset := 0; ; offset += batchSize {
		variables["offset"] = graphql.Int(offset)

		raw, err := impl.graphql.ExecRaw(ctx, query, variables, nil)
		if err != nil {
			return nil, fmt.Errorf("graphql.ExecRaw error: %w", err)
		}

		var result struct {
			CurrentTokenOwnerships []struct {
				Creator         string            `json:"creator_address"`
				Collection      string            `json:"collection_name"`
				Name            string            `json:"name"`
				PropertyVersion models.Uint64     `json:"property_version"`
				Amount          models.Uint64     `json:"amount"`
				TokenProperties map[string]string `json:"token_properties"`
				TokenDataIDHash string            `json:"token_data_id_hash"`
			} `json:"current_token_ownerships"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, fmt.Errorf("json.Unmarshal error: %w", err)
		}

		for _, t := range result.CurrentTokenOwnerships {
			tokens = append(tokens, models.Token{
				ID: models.TokenID{
					TokenDataID: models.TokenDataID{
						Hash:       t.TokenDataIDHash,
						Creator:    t.Creator,
						Collection: t.Collection,
						Name:       t.Name,
					},
					PropertyVersion: t.PropertyVersion,
				},
				Amount:         t.Amount,
				JSONProperties: t.TokenProperties,
			})
		}

		if len(result.CurrentTokenOwnerships) < batchSize {
			break
		}
	}

	return tokens, nil
}

// ListAccountTokens gets aptos tokens of an account by indexer graphql api. Returns a list of tokens of version v1 and v2.
func (impl *TokenClientImpl) ListAccountTokensV2(ctx context.Context, owners ...models.AccountAddress) ([]models.TokenV2, error) {
	if len(owners) == 0 {
		return nil, nil
	}

	const batchSize = 100

	query := `
	query CurrentTokens($owner_addresses: [String!], $offset: Int, $limit: Int) {
		current_token_ownerships_v2(
			where: {owner_address: {_in: $owner_addresses}, amount: {_gt: "0"}}
			order_by: {last_transaction_version: asc}
			offset: $offset
			limit: $limit
			) {
				amount
				current_token_data {
					current_collection {
						creator_address
						collection_name
						current_supply
						max_supply
					}
					token_name
					token_uri
					description
				}
				owner_address
				property_version_v1
				is_soulbound_v2
				token_data_id
				token_standard
			}
		}
	`

	addresses := make([]graphql.String, 0, len(owners))
	for i := range owners {
		addresses = append(addresses, graphql.String(owners[i].ToHex()))
	}
	variables := map[string]interface{}{
		"owner_addresses": addresses,
		"limit":           batchSize,
	}

	var tokens []models.TokenV2
	for offset := 0; ; offset += batchSize {
		variables["offset"] = graphql.Int(offset)

		raw, err := impl.graphql.ExecRaw(ctx, query, variables, nil)
		if err != nil {
			return nil, fmt.Errorf("graphql.ExecRaw error: %w", err)
		}

		var result struct {
			CurrentTokenOwnershipsV2 []struct {
				Amount            models.Uint64 `json:"amount"`
				OwnerAddress      string        `json:"owner_address"`
				PropertyVersionV1 models.Uint64 `json:"property_version_v1"`
				IsSoulboundV2     bool          `json:"is_soulbound_v2"`
				TokenDataID       string        `json:"token_data_id"`
				TokenStandard     string        `json:"token_standard"`
				CurrentTokenData  struct {
					CurrentCollection struct {
						CreatorAddress string         `json:"creator_address"`
						CollectionName string         `json:"collection_name"`
						MaxSupply      *models.Uint64 `json:"max_supply"`
						CurrentSupply  models.Uint64  `json:"current_supply"`
					} `json:"current_collection"`
					TokenName   string `json:"token_name"`
					TokenURI    string `json:"token_uri"`
					Description string `json:"description"`
				} `json:"current_token_data"`
			} `json:"current_token_ownerships_v2"`
		}

		fmt.Println("raw", string(raw))
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, fmt.Errorf("json.Unmarshal error: %w", err)
		}
		fmt.Printf("result: %+v\n", result)

		for i := range result.CurrentTokenOwnershipsV2 {
			tokens = append(tokens, models.TokenV2{
				ID:                result.CurrentTokenOwnershipsV2[i].TokenDataID,
				Name:              result.CurrentTokenOwnershipsV2[i].CurrentTokenData.TokenName,
				Description:       result.CurrentTokenOwnershipsV2[i].CurrentTokenData.Description,
				URI:               result.CurrentTokenOwnershipsV2[i].CurrentTokenData.TokenURI,
				Standard:          result.CurrentTokenOwnershipsV2[i].TokenStandard,
				OwnerAddress:      result.CurrentTokenOwnershipsV2[i].OwnerAddress,
				Amount:            result.CurrentTokenOwnershipsV2[i].Amount,
				CollectionName:    result.CurrentTokenOwnershipsV2[i].CurrentTokenData.CurrentCollection.CollectionName,
				CreatorAddress:    result.CurrentTokenOwnershipsV2[i].CurrentTokenData.CurrentCollection.CreatorAddress,
				Maximum:           result.CurrentTokenOwnershipsV2[i].CurrentTokenData.CurrentCollection.MaxSupply,
				Supply:            result.CurrentTokenOwnershipsV2[i].CurrentTokenData.CurrentCollection.CurrentSupply,
				PropertyVersionV1: result.CurrentTokenOwnershipsV2[i].PropertyVersionV1,
				IsSoulboundV2:     result.CurrentTokenOwnershipsV2[i].IsSoulboundV2,
			})
		}

		if len(result.CurrentTokenOwnershipsV2) < batchSize {
			break
		}
	}

	return tokens, nil
}
