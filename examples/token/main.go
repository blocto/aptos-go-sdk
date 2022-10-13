package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/portto/aptos-go-sdk/client"
	"github.com/portto/aptos-go-sdk/crypto"
	"github.com/portto/aptos-go-sdk/models"
)

const DevnetChainID = 33

const CollectionName = "Aptos"
const TokenName = "Aptos Token"

var aptosClient client.AptosClient
var tokenClient client.TokenClient

var addr0x1 models.AccountAddress
var aptosCoinTypeTag models.TypeTag

var faucetAdminSeed []byte
var faucetAdmin models.SingleSigner
var faucetAdminAddress string
var faucetAdminAddr models.AccountAddress

var ctx = context.Background()

func init() {
	var err error

	aptosClient = client.NewAptosClient("https://fullnode.devnet.aptoslabs.com")
	tokenClient, err = client.NewTokenClient(aptosClient)
	if err != nil {
		panic(err)
	}

	addr0x1, _ = models.HexToAccountAddress("0x1")

	aptosCoinTypeTag = models.TypeTagStruct{
		Address: addr0x1,
		Module:  "aptos_coin",
		Name:    "AptosCoin",
	}

	faucetAdminSeed, _ = hex.DecodeString("784bc4d62c5e96b42addcbee3e5ccc0f7641fa82e9a3462d9a34d06e474274fe")
	faucetAdmin = models.NewSingleSigner(ed25519.NewKeyFromSeed(faucetAdminSeed))
	faucetAdminAddress = "86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850"
	faucetAdminAddr, _ = models.HexToAccountAddress(faucetAdminAddress)
}

func main() {
	txHash, addr, seed := createAccount()
	if err := aptosClient.WaitForTransaction(ctx, txHash); err != nil {
		panic(err)
	}

	txHash = faucet(addr, 500000)
	if err := aptosClient.WaitForTransaction(ctx, txHash); err != nil {
		panic(err)
	}

	creator := models.NewSingleSigner(ed25519.NewKeyFromSeed(seed))

	ctx := context.Background()

	hash, err := tokenClient.CreateCollection(ctx, creator, client.CreateCollectionRequest{
		Name:        CollectionName,
		Description: "Blocto",
		URI:         "https://blocto.app",
		Maximum:     25600,
		MutateConfig: client.CollectionMutabilityConfig{
			Description: true,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("create collection hash: %v\n", hash)

	if err := aptosClient.WaitForTransaction(ctx, hash); err != nil {
		panic(err)
	}

	hash, err = tokenClient.CreateToken(ctx, creator, client.CreateTokenRequest{
		Collection:  CollectionName,
		Name:        TokenName,
		Description: "Blocto",
		Supply:      2,
		URI:         "https://blocto.app",
		MutateConfig: client.TokenMutabilityConfig{
			Description: true,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("create token hash: %v\n", hash)

	if err := aptosClient.WaitForTransaction(ctx, hash); err != nil {
		panic(err)
	}

	hash, err = tokenClient.MintToken(ctx, creator, client.MintTokenRequest{
		Creator:    creator.AccountAddress,
		Collection: CollectionName,
		TokenName:  TokenName,
		Amount:     3,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("mint token hash: %v\n", hash)

	if err := aptosClient.WaitForTransaction(ctx, hash); err != nil {
		panic(err)
	}

	hash, err = tokenClient.OfferToken(ctx, creator, client.OfferTokenRequest{
		Receiver:   faucetAdminAddr,
		Creator:    creator.AccountAddress,
		Collection: CollectionName,
		TokenName:  TokenName,
		Amount:     1,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("offer token hash: %v\n", hash)

	if err := aptosClient.WaitForTransaction(ctx, hash); err != nil {
		panic(err)
	}

	hash, err = tokenClient.ClaimToken(ctx, faucetAdmin, client.ClaimTokenRequest{
		Sender:     creator.AccountAddress,
		Creator:    creator.AccountAddress,
		Collection: CollectionName,
		TokenName:  TokenName,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("claim token hash: %v\n", hash)

	if err := aptosClient.WaitForTransaction(ctx, hash); err != nil {
		panic(err)
	}

	collectionData, err := tokenClient.GetCollectionData(ctx, creator.AccountAddress, CollectionName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("collection data: %+v\n", collectionData)

	tokenData, err := tokenClient.GetTokenData(ctx, creator.AccountAddress, CollectionName, TokenName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("token data: %+v\n", tokenData)

	token, err := tokenClient.GetToken(ctx, creator.AccountAddress, client.TokenID{
		TokenDataID: client.TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: CollectionName,
			Name:       TokenName,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("token: %+v\n", token)

	token, err = tokenClient.GetToken(ctx, faucetAdminAddr, client.TokenID{
		TokenDataID: client.TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: CollectionName,
			Name:       TokenName,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("token: %+v\n", token)
}

func createAccount() (txHash string, addr models.AccountAddress, seed []byte) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	authKey := crypto.SingleSignerAuthKey(pub)

	accountInfo, err := aptosClient.GetAccount(ctx, faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}

	addr0x1, _ := models.HexToAccountAddress("0x1")
	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(models.EntryFunctionPayload{
			Module: models.Module{
				Address: addr0x1,
				Name:    "aptos_account",
			},
			Function:  "create_account",
			Arguments: []interface{}{authKey},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1000)).
		SetMaxGasAmount(uint64(1000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signer := models.NewSingleSigner(ed25519.NewKeyFromSeed(faucetAdminSeed))
	err = signer.Sign(&tx).Error()
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("create account tx hash:", rawTx.Hash)
	return rawTx.Hash, authKey, priv.Seed()
}

func faucet(address models.AccountAddress, amount uint64) string {
	accountInfo, err := aptosClient.GetAccount(ctx, faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}

	if err != nil {
		panic(err)
	}
	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(getTransferPayload(address, amount)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(500)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	err = faucetAdmin.Sign(&tx).Error()
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("faucet tx hash:", rawTx.Hash)
	return rawTx.Hash
}

func getTransferPayload(to models.AccountAddress, amount uint64) models.TransactionPayload {
	return models.EntryFunctionPayload{
		Module: models.Module{
			Address: addr0x1,
			Name:    "coin",
		},
		Function:      "transfer",
		TypeArguments: []models.TypeTag{aptosCoinTypeTag},
		Arguments:     []interface{}{to, amount},
	}
}
