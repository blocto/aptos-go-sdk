package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/portto/aptos-go-sdk/client"
	"github.com/portto/aptos-go-sdk/crypto"
	"github.com/portto/aptos-go-sdk/transactions"
)

var api client.API
var faucetAdminSeed []byte
var faucetAdminAddress string

func init() {
	api = client.New("https://fullnode.devnet.aptoslabs.com/v1")
	// please set up the account address & seed which has enough balance
	faucetAdminSeed, _ = hex.DecodeString("784bc4d62c5e96b42addcbee3e5ccc0f7641fa82e9a3462d9a34d06e474274fe")
	faucetAdminAddress = "86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850"
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
	invokeMultiAgentScriptPayload("multi_agent_set_message", []string{}, []interface{}{
		hex.EncodeToString([]byte("hi~ aptos")),
	})
	fmt.Println("=====")
	waitForTxConfirmed()
	invokeMultiAgentScriptPayload("multi_agent_rotate_authentication_key", []string{}, []interface{}{
		hex.EncodeToString(make([]byte, 32)),
	})
	fmt.Println("=====")
	waitForTxConfirmed()
	invokeMultiAgentScriptPayload("multi_agent_transfer", []string{
		"0x1::aptos_coin::AptosCoin",
	}, []interface{}{
		faucetAdminAddress,
		"1",
	}, "1")
	fmt.Println("=====")
	transferTxWeightedMultiED25519()
	fmt.Println("=====")
}

func waitForTxConfirmed() {
	time.Sleep(15 * time.Second)
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
	faucet(address, "50")
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	newAuthKey := crypto.SingleSignerAuthKey(newPub)
	tx := transactions.Transaction{}
	err = tx.SetSender(address).
		SetPayload("script_function_payload",
			[]string{},
			[]interface{}{
				hex.EncodeToString(newAuthKey[:]),
			}, client.ScriptFunctionPayload{
				Function: "0x1::account::rotate_authentication_key",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(50)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetSignature("ed25519_signature", client.ED25519Signature{
		PublicKey: hex.EncodeToString(priv.Public().(ed25519.PublicKey)[:]),
		Signature: hex.EncodeToString(signature),
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("replace auth key tx hash:", rawTx.Hash)
}

func transferTxMultiED25519() {
	authKey, seeds := createAccountTx(2)
	address := hex.EncodeToString(authKey[:])
	waitForTxConfirmed()
	faucet(address, "100")
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	tx := transactions.Transaction{}
	err = tx.SetSender(address).
		SetPayload("script_function_payload",
			[]string{"0x1::aptos_coin::AptosCoin"},
			[]interface{}{
				faucetAdminAddress,
				"1",
			}, client.ScriptFunctionPayload{
				Function: "0x1::coin::transfer",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	key1, _ := hex.DecodeString(seeds[0])
	key2, _ := hex.DecodeString(seeds[1])
	priv1 := ed25519.NewKeyFromSeed(key1)
	priv2 := ed25519.NewKeyFromSeed(key2)
	signature := ed25519.Sign(priv2, msgBytes)
	err = tx.SetSignature("multi_ed25519_signature", client.MultiED25519Signature{
		PublicKeys: []string{hex.EncodeToString(priv1.Public().(ed25519.PublicKey)), hex.EncodeToString(priv2.Public().(ed25519.PublicKey))},
		Signatures: []string{hex.EncodeToString(signature)},
		Threshold:  1,
		Bitmap:     "40000000",
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

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

	tx := transactions.Transaction{}
	err = tx.SetSender(faucetAdminAddress).
		SetPayload("script_function_payload",
			[]string{},
			[]interface{}{
				hex.EncodeToString(authKey[:]),
			}, client.ScriptFunctionPayload{
				Function: "0x1::account::create_account",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(1000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetSignature("ed25519_signature", client.ED25519Signature{
		PublicKey: hex.EncodeToString(faucetAdminPriv.Public().(ed25519.PublicKey)),
		Signature: hex.EncodeToString(signature),
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("create account tx hash:", rawTx.Hash)
	return authKey, seeds
}

func faucet(address string, amount string) {
	accountInfo, err := api.GetAccount(faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	priv := ed25519.NewKeyFromSeed(faucetAdminSeed)
	tx := transactions.Transaction{}
	err = tx.SetSender(faucetAdminAddress).
		SetPayload("script_function_payload",
			[]string{"0x1::aptos_coin::AptosCoin"},
			[]interface{}{
				address,
				amount,
			}, client.ScriptFunctionPayload{
				Function: "0x1::coin::transfer",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(500)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetSignature("ed25519_signature", client.ED25519Signature{
		PublicKey: hex.EncodeToString(priv.Public().(ed25519.PublicKey)[:]),
		Signature: hex.EncodeToString(signature),
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

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

	tx := transactions.Transaction{}
	err = tx.SetSender(sender).
		SetPayload("script_function_payload",
			[]string{},
			[]interface{}{
				hex.EncodeToString([]byte("aptos-is-goooooood!")),
			}, client.ScriptFunctionPayload{
				Function: "0x86e4d830197448f975b748f69bd1b3b6d219a07635269a0b4e7f27966771e850::message_multi_agent::set_message",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(senderInfo.SequenceNumber).
		SetSecondarySigners([]string{
			hex.EncodeToString(authKey[:]),
		}).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
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
	err = tx.SetSignature("multi_agent_signature", client.MultiAgentSignature{
		Sender: struct {
			Type string `json:"type"`
			client.ED25519Signature
			client.MultiED25519Signature
		}{
			Type: "ed25519_signature",
			ED25519Signature: client.ED25519Signature{
				PublicKey: hex.EncodeToString(senderPriv.Public().(ed25519.PublicKey)),
				Signature: hex.EncodeToString(senderSignature),
			},
		},
		SecondarySignerAddresses: []string{
			hex.EncodeToString(authKey[:]),
		},
		SecondarySigners: []struct {
			Type string `json:"type"`
			client.ED25519Signature
			client.MultiED25519Signature
		}{
			{
				Type: "multi_ed25519_signature",
				MultiED25519Signature: client.MultiED25519Signature{
					PublicKeys: []string{hex.EncodeToString(priv1.Public().(ed25519.PublicKey)), hex.EncodeToString(priv2.Public().(ed25519.PublicKey))},
					Signatures: []string{hex.EncodeToString(signature)},
					Threshold:  1,
					Bitmap:     "40000000",
				},
			},
		},
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

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
	faucet(address, "50")
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	tx := transactions.Transaction{}
	err = tx.SetSender(address).
		SetPayload("script_payload",
			[]string{},
			[]interface{}{
				hex.EncodeToString([]byte("hi script payload~")),
			}, client.ScriptPayload{
				Code: client.Code{
					Bytecode: hex.EncodeToString(fileBytes),
				},
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(50)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetSignature("ed25519_signature", client.ED25519Signature{
		PublicKey: hex.EncodeToString(priv.Public().(ed25519.PublicKey)[:]),
		Signature: hex.EncodeToString(signature),
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("test script payload tx hash:", rawTx.Hash)
}

func invokeMultiAgentScriptPayload(scriptName string, typeArgs []string, args []interface{}, faucetAmount ...string) {
	fileBytes, err := os.ReadFile(fmt.Sprintf("./examples/transactions/scripts/%s.mv", scriptName))
	if err != nil {
		panic(err)
	}

	authKey, seeds := createAccountTx(2)
	waitForTxConfirmed()

	if len(faucetAmount) > 0 {
		faucet(hex.EncodeToString(authKey[:]), faucetAmount[0])
		waitForTxConfirmed()
	}

	sender := faucetAdminAddress
	senderInfo, err := api.GetAccount(sender)
	if err != nil {
		panic(err)
	}

	tx := transactions.Transaction{}
	err = tx.SetSender(sender).
		SetPayload("script_payload",
			typeArgs,
			args,
			client.ScriptPayload{
				Code: client.Code{
					Bytecode: hex.EncodeToString(fileBytes),
				},
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(senderInfo.SequenceNumber).
		SetSecondarySigners([]string{
			hex.EncodeToString(authKey[:]),
		}).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
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
	err = tx.SetSignature("multi_agent_signature", client.MultiAgentSignature{
		Sender: struct {
			Type string `json:"type"`
			client.ED25519Signature
			client.MultiED25519Signature
		}{
			Type: "ed25519_signature",
			ED25519Signature: client.ED25519Signature{
				PublicKey: hex.EncodeToString(senderPriv.Public().(ed25519.PublicKey)),
				Signature: hex.EncodeToString(senderSignature),
			},
		},
		SecondarySignerAddresses: []string{
			hex.EncodeToString(authKey[:]),
		},
		SecondarySigners: []struct {
			Type string `json:"type"`
			client.ED25519Signature
			client.MultiED25519Signature
		}{
			{
				Type: "multi_ed25519_signature",
				MultiED25519Signature: client.MultiED25519Signature{
					PublicKeys: []string{hex.EncodeToString(priv1.Public().(ed25519.PublicKey)), hex.EncodeToString(priv2.Public().(ed25519.PublicKey))},
					Signatures: []string{hex.EncodeToString(signature)},
					Threshold:  1,
					Bitmap:     "40000000",
				},
			},
		},
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

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

	tx := transactions.Transaction{}
	err = tx.SetSender(faucetAdminAddress).
		SetPayload("script_function_payload",
			[]string{},
			[]interface{}{
				hex.EncodeToString(authKey[:]),
			}, client.ScriptFunctionPayload{
				Function: "0x1::account::create_account",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(1000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetSignature("ed25519_signature", client.ED25519Signature{
		PublicKey: hex.EncodeToString(faucetAdminPriv.Public().(ed25519.PublicKey)),
		Signature: hex.EncodeToString(signature),
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("create account tx hash:", rawTx.Hash)
	return authKey, seeds
}

func transferTxWeightedMultiED25519() {
	authKey, seeds := createWeightAccountTx()
	address := hex.EncodeToString(authKey[:])
	waitForTxConfirmed()
	faucet(address, "100")
	waitForTxConfirmed()

	accountInfo, err := api.GetAccount(address)
	if err != nil {
		panic(err)
	}

	tx := transactions.Transaction{}
	err = tx.SetSender(address).
		SetPayload("script_function_payload",
			[]string{"0x1::aptos_coin::AptosCoin"},
			[]interface{}{
				faucetAdminAddress,
				"1",
			}, client.ScriptFunctionPayload{
				Function: "0x1::coin::transfer",
			}).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(1)).
		SetMaxGasAmount(uint64(100)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signingMsg, err := api.EncodeSubmission(tx.Transaction)
	if err != nil {
		panic(err)
	}

	err = tx.SetSigningMessage(signingMsg.Message).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := hex.DecodeString(signingMsg.Message[2:])
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
	err = tx.SetSignature("multi_ed25519_signature", client.MultiED25519Signature{
		PublicKeys: []string{hex.EncodeToString(priv1.Public().(ed25519.PublicKey)), hex.EncodeToString(priv2.Public().(ed25519.PublicKey)),
			hex.EncodeToString(priv3.Public().(ed25519.PublicKey)), hex.EncodeToString(priv3.Public().(ed25519.PublicKey))},
		Signatures: []string{hex.EncodeToString(signature), hex.EncodeToString(signature)},
		Threshold:  2,
		Bitmap:     "30000000",
	}).Error()
	if err != nil {
		panic(err)
	}

	txForSimulate := tx.TxForSimulate()
	_, err = api.SimulateTransaction(txForSimulate)
	if err != nil {
		panic(err)
	}

	rawTx, err := api.SubmitTransaction(tx.Transaction)
	if err != nil {
		panic(err)
	}

	fmt.Println("transfer tx hash:", rawTx.Hash)
}
