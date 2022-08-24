package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/portto/aptos-go-sdk/client"
	"github.com/portto/aptos-go-sdk/crypto"
	"github.com/portto/aptos-go-sdk/models"
)

const DevnetChainID = 24

var api client.API
var faucetAdminSeed []byte
var faucetAdminAddress string
var faucetAdminAddr models.AccountAddress
var addr0x1 models.AccountAddress
var aptosCoinTypeTag models.TypeTag

func init() {
	api = client.New("https://fullnode.devnet.aptoslabs.com")
	// please set up the account address & seed which has enough balance
	faucetAdminSeed, _ = hex.DecodeString("784bc4d62c5e96b42addcbee3e5ccc0f7641fa82e9a3462d9a34d06e474274fe")
	faucetAdminAddress = "86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850"
	faucetAdminAddr, _ = models.HexToAccountAddress(faucetAdminAddress)
	addr0x1, _ = models.HexToAccountAddress("0x01")
	aptosCoinTypeTag = models.TypeTagStruct{
		Address: addr0x1,
		Module:  "aptos_coin",
		Name:    "AptosCoin",
	}
}

func main() {
	replaceAuthKey()
	fmt.Println("=====")
	transferTxMultiED25519()
	fmt.Println("=====")
	invokeMultiAgent()
	fmt.Println("=====")
	waitForTxConfirmed()
	invokeScriptPayload()
	fmt.Println("=====")
	invokeMultiAgentScriptPayload("multi_agent_set_message", nil,
		[]models.TransactionArgument{
			models.TxArgU8Vector{Bytes: []byte("hi~ aptos")},
		})
	fmt.Println("=====")
	waitForTxConfirmed()
	invokeMultiAgentScriptPayload("multi_agent_rotate_authentication_key", nil,
		[]models.TransactionArgument{
			models.TxArgU8Vector{Bytes: make([]byte, 32)},
		})
	fmt.Println("=====")
	waitForTxConfirmed()
	invokeMultiAgentScriptPayload("multi_agent_coin_transfer",
		[]models.TypeTag{
			aptosCoinTypeTag,
		},
		[]models.TransactionArgument{
			models.TxArgAddress{Addr: faucetAdminAddr},
			models.TxArgU64{U64: 1},
		}, 1)
	fmt.Println("=====")
	waitForTxConfirmed()
	transferTxWeightedMultiED25519()
	fmt.Println("=====")
	waitForTxConfirmed()
	moonCoinAddr, _ := models.HexToAccountAddress("0x86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850")
	invokeMultiAgentScriptPayload("multi_agent_coin_register", []models.TypeTag{
		models.TypeTagStruct{
			Address: moonCoinAddr,
			Module:  "moon_coin",
			Name:    "MoonCoin",
		},
	}, []models.TransactionArgument{})
	fmt.Println("=====")
}

func waitForTxConfirmed() {
	time.Sleep(5 * time.Second)
}

func replaceAuthKey() {
	authKey, seeds := createAccountTx(1)
	key, err := hex.DecodeString(seeds[0])
	if err != nil {
		panic(err)
	}

	priv := ed25519.NewKeyFromSeed(key)
	address := hex.EncodeToString(authKey[:])
	newPub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	waitForTxConfirmed()
	faucet(authKey, 50)
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	newAuthKey := crypto.SingleSignerAuthKey(newPub)
	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(models.EntryFunctionPayload{
			Module: models.Module{
				Address: addr0x1,
				Name:    "account",
			},
			Function: "rotate_authentication_key",
			Arguments: []interface{}{
				newAuthKey[:],
			},
		},
		).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(50)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: priv.Public().(ed25519.PublicKey),
		Signature: signature,
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("           computed hash:", hash)
	fmt.Println("replace auth key tx hash:", rawTx.Hash)
}

func transferTxMultiED25519() {
	authKey, seeds := createAccountTx(2)
	address := hex.EncodeToString(authKey[:])
	waitForTxConfirmed()
	faucet(authKey, 100)
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	addr, _ := models.HexToAccountAddress(faucetAdminAddress)

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(getTransferPayload(addr, 1)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	key1, _ := hex.DecodeString(seeds[0])
	key2, _ := hex.DecodeString(seeds[1])
	priv1 := ed25519.NewKeyFromSeed(key1)
	priv2 := ed25519.NewKeyFromSeed(key2)
	signature := ed25519.Sign(priv2, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorMultiEd25519{
		PublicKeys: []models.PublicKey{
			priv1.Public().(ed25519.PublicKey),
			priv2.Public().(ed25519.PublicKey),
		},
		Threshold:  1,
		Signatures: []models.Signature{signature},
		Bitmap:     [4]byte{0x40, 0, 0, 0},
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("   computed hash:", hash)

	fmt.Println("transfer tx hash:", rawTx.Hash)
}

func createAccountTx(keyNum int) (authKey [32]byte, seeds []string) {
	faucetAdminPriv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	if keyNum == 1 {
		pub, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			panic(err)
		}

		seeds = append(seeds, hex.EncodeToString(priv.Seed()))
		authKey = crypto.SingleSignerAuthKey(pub)
	} else if keyNum > 1 {
		var publicKeys [][]byte
		for i := 0; i < keyNum; i++ {
			pub, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				panic(err)
			}

			publicKeys = append(publicKeys, pub)
			seeds = append(seeds, hex.EncodeToString(priv.Seed()))
		}
		authKey = crypto.MultiSignerAuthKey(1, publicKeys...)
	} else {
		panic("keyNum is zero")
	}

	accountInfo, err := api.GetAccount(faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}

	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(models.EntryFunctionPayload{
			Module: models.Module{
				Address: addr0x1,
				Name:    "account",
			},
			Function:  "create_account",
			Arguments: []interface{}{authKey},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(1000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: faucetAdminPriv.Public().(ed25519.PublicKey),
		Signature: signature,
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("         computed hash:", hash)

	fmt.Println("create account tx hash:", rawTx.Hash)
	return authKey, seeds
}

func faucet(address models.AccountAddress, amount uint64) {
	accountInfo, err := api.GetAccount(faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	priv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	tx := models.Transaction{}

	if err != nil {
		panic(err)
	}
	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(getTransferPayload(address, amount)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(500)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: priv.Public().(ed25519.PublicKey),
		Signature: signature,
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println(" computed hash:", hash)

	fmt.Println("faucet tx hash:", rawTx.Hash)
}

func invokeMultiAgent() {
	authKey, seeds := createAccountTx(2)
	waitForTxConfirmed()
	sender := faucetAdminAddress
	senderInfo, err := api.GetAccount(sender)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}

	addr, _ := models.HexToAccountAddress("0x86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850")
	err = tx.SetChainID(DevnetChainID).
		SetSender(sender).
		SetPayload(models.EntryFunctionPayload{
			Module: models.Module{
				Address: addr,
				Name:    "message_multi_agent_1",
			},
			Function:  "set_message",
			Arguments: []interface{}{"aptos-is-goooooood!"},
		},
		).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(senderInfo.SequenceNumber).
		SetSecondarySigners([]models.AccountAddress{authKey}).
		Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	key1, _ := hex.DecodeString(seeds[0])
	key2, _ := hex.DecodeString(seeds[1])
	priv1 := ed25519.NewKeyFromSeed(key1)
	priv2 := ed25519.NewKeyFromSeed(key2)
	signature := ed25519.Sign(priv2, msgBytes)
	senderPriv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	senderSignature := ed25519.Sign(senderPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorMultiAgent{
		Sender: models.AccountAuthenticatorEd25519{
			PublicKey: senderPriv.Public().(ed25519.PublicKey),
			Signature: senderSignature,
		},
		SecondarySignerAddresses: []models.AccountAddress{authKey},
		SecondarySigners: []models.AccountAuthenticator{
			models.AccountAuthenticatorMultiEd25519{
				PublicKeys: []models.PublicKey{
					priv1.Public().(ed25519.PublicKey),
					priv2.Public().(ed25519.PublicKey),
				},
				Threshold:  1,
				Signatures: []models.Signature{signature},
				Bitmap:     [4]byte{0x40, 0, 0, 0},
			},
		},
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("     computed hash:", hash)
	fmt.Println("multiAgent tx hash:", rawTx.Hash)
}

func invokeScriptPayload() {
	fileBytes, err := os.ReadFile("./examples/transactions/scripts/set_message.mv")
	if err != nil {
		panic(err)
	}

	authKey, seeds := createAccountTx(1)
	key, err := hex.DecodeString(seeds[0])
	if err != nil {
		panic(err)
	}

	priv := ed25519.NewKeyFromSeed(key)
	address := hex.EncodeToString(authKey[:])
	waitForTxConfirmed()
	faucet(authKey, 50)
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(models.ScriptPayload{
			Code: fileBytes,
			Arguments: []models.TransactionArgument{
				models.TxArgU8Vector{Bytes: []byte("hi script payload~")},
			},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(50)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		fmt.Printf("%+v\n", tx.UserTransaction)
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: priv.Public().(ed25519.PublicKey),
		Signature: signature,
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("              computed hash:", hash)
	fmt.Println("test script payload tx hash:", rawTx.Hash)
}

func invokeMultiAgentScriptPayload(scriptName string, typeArgs []models.TypeTag, args []models.TransactionArgument, faucetAmount ...uint64) {
	fileBytes, err := os.ReadFile(fmt.Sprintf("./examples/transactions/scripts/%s.mv", scriptName))
	if err != nil {
		panic(err)
	}

	authKey, seeds := createAccountTx(2)
	waitForTxConfirmed()

	if len(faucetAmount) > 0 {
		faucet(authKey, faucetAmount[0])
		waitForTxConfirmed()
	}

	sender := faucetAdminAddress
	senderInfo, err := api.GetAccount(sender)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(sender).
		SetPayload(models.ScriptPayload{
			Code:          fileBytes,
			TypeArguments: typeArgs,
			Arguments:     args,
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(senderInfo.SequenceNumber).
		SetSecondarySigners([]models.AccountAddress{authKey}).
		Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	key1, _ := hex.DecodeString(seeds[0])
	key2, _ := hex.DecodeString(seeds[1])
	priv1 := ed25519.NewKeyFromSeed(key1)
	priv2 := ed25519.NewKeyFromSeed(key2)
	signature := ed25519.Sign(priv2, msgBytes)
	senderPriv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	senderSignature := ed25519.Sign(senderPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorMultiAgent{
		Sender: models.AccountAuthenticatorEd25519{
			PublicKey: senderPriv.Public().(ed25519.PublicKey),
			Signature: senderSignature,
		},
		SecondarySignerAddresses: []models.AccountAddress{authKey},
		SecondarySigners: []models.AccountAuthenticator{
			models.AccountAuthenticatorMultiEd25519{
				PublicKeys: []models.PublicKey{
					priv1.Public().(ed25519.PublicKey),
					priv2.Public().(ed25519.PublicKey),
				},
				Threshold:  1,
				Signatures: []models.Signature{signature},
				Bitmap:     [4]byte{0x40, 0, 0, 0},
			},
		},
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("                         computed hash:", hash)
	fmt.Println("test multiAgent script payload tx hash:", rawTx.Hash)
}

func createWeightAccountTx() (authKey [32]byte, seeds []string) {
	faucetAdminPriv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	var publicKeys [][]byte
	key1Pub, key1Priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	seeds = append(seeds, hex.EncodeToString(key1Priv.Seed()))
	publicKeys = append(publicKeys, key1Pub)

	key2Pub, key2Priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	seeds = append(seeds, hex.EncodeToString(key2Priv.Seed()))
	publicKeys = append(publicKeys, key2Pub)

	key3Pub, key3Priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	seeds = append(seeds, hex.EncodeToString(key3Priv.Seed()))
	publicKeys = append(publicKeys, key3Pub)
	publicKeys = append(publicKeys, key3Pub)

	authKey = crypto.MultiSignerAuthKey(2, publicKeys...)
	accountInfo, err := api.GetAccount(faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(models.EntryFunctionPayload{
			Module: models.Module{
				Address: addr0x1,
				Name:    "account",
			},
			Function:  "create_account",
			Arguments: []interface{}{authKey},
		}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(1000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: faucetAdminPriv.Public().(ed25519.PublicKey),
		Signature: signature,
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("         computed hash:", hash)

	fmt.Println("create account tx hash:", rawTx.Hash)
	return authKey, seeds
}

func transferTxWeightedMultiED25519() {
	authKey, seeds := createWeightAccountTx()
	address := hex.EncodeToString(authKey[:])
	waitForTxConfirmed()
	faucet(authKey, 100)
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(getTransferPayload(faucetAdminAddr, 1)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(hex.EncodeToString(msgBytes)).Error()
	if err != nil {
		panic(err)
	}

	key1, _ := hex.DecodeString(seeds[0])
	key2, _ := hex.DecodeString(seeds[1])
	key3, _ := hex.DecodeString(seeds[2])
	priv1 := ed25519.NewKeyFromSeed(key1)
	priv2 := ed25519.NewKeyFromSeed(key2)
	priv3 := ed25519.NewKeyFromSeed(key3)
	signature := ed25519.Sign(priv3, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorMultiEd25519{
		PublicKeys: []models.PublicKey{
			priv1.Public().(ed25519.PublicKey),
			priv2.Public().(ed25519.PublicKey),
			priv3.Public().(ed25519.PublicKey),
			priv3.Public().(ed25519.PublicKey),
		},
		Threshold:  2,
		Signatures: []models.Signature{signature, signature},
		Bitmap:     [4]byte{0x30, 0, 0, 0},
	}).Error()
	if err != nil {
		panic(err)
	}

	txReq := tx.ToRequest()
	_, err = api.SimulateTransaction(txReq.ForSimulate())
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(txReq)
	if err != nil {
		panic(err)
	}

	hash, err := tx.GetHash()
	if err != nil {
		panic(err)
	}
	fmt.Println("   computed hash:", hash)
	fmt.Println("transfer tx hash:", rawTx.Hash)
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
