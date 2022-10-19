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

const DevnetChainID = 2

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

	aptosClient = client.NewAptosClient("https://fullnode.testnet.aptoslabs.com")
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

	faucetAdminSeed, _ = hex.DecodeString("3a645d4b735c6ad68d727955dd89e5c1d5546059362f7a96a496a26381ae8221")
	faucetAdmin = models.NewSingleSigner(ed25519.NewKeyFromSeed(faucetAdminSeed))
	faucetAdminAddress = "2d0f23232bdcd3862c2f65989063415219c4486fcc68c542b3d86ec18de4c9e6"
	faucetAdminAddr, _ = models.HexToAccountAddress(faucetAdminAddress)
}

func main() {
	txHash, addr, seed := createAccount()
	if err := aptosClient.WaitForTransaction(ctx, txHash); err != nil {
		panic(err)
	}

	txHash = faucet(addr, 5000000)
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
		MutateConfig: models.TokenMutabilityConfig{
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

	token, err := tokenClient.GetToken(ctx, creator.AccountAddress, models.TokenID{
		TokenDataID: models.TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: CollectionName,
			Name:       TokenName,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("token: %+v\n", token)

	token, err = tokenClient.GetToken(ctx, faucetAdminAddr, models.TokenID{
		TokenDataID: models.TokenDataID{
			Creator:    creator.PrefixZeroTrimmedHex(),
			Collection: CollectionName,
			Name:       TokenName,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("token: %+v\n", token)

	tokens, err := tokenClient.ListAccountTokens(ctx, creator.AccountAddress)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", tokens)
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
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
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
		SetMaxGasAmount(uint64(5000)).
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
