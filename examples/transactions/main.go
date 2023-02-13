package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/the729/lcs"

	"github.com/portto/aptos-go-sdk/client"
	"github.com/portto/aptos-go-sdk/crypto"
	"github.com/portto/aptos-go-sdk/models"
)

const DevnetChainID = 34

var aptosClient client.AptosClient
var faucetAdminSeed []byte
var faucetAdminAddress string
var faucetAdminAddr models.AccountAddress
var addr0x1 models.AccountAddress
var aptosAccountModule models.Module
var accountModule models.Module
var aptosCoinTypeTag models.TypeTag
var ctx = context.Background()

func init() {
	aptosClient = client.NewAptosClient("https://fullnode.devnet.aptoslabs.com")
	// please set up the account address & seed which has enough balance
	faucetAdminSeed, _ = hex.DecodeString("3a645d4b735c6ad68d727955dd89e5c1d5546059362f7a96a496a26381ae8221")
	faucetAdminAddress = "2d0f23232bdcd3862c2f65989063415219c4486fcc68c542b3d86ec18de4c9e6"
	faucetAdminAddr, _ = models.HexToAccountAddress(faucetAdminAddress)
	addr0x1, _ = models.HexToAccountAddress("0x1")
	aptosAccountModule = models.Module{
		Address: addr0x1,
		Name:    "aptos_account",
	}
	accountModule = models.Module{
		Address: addr0x1,
		Name:    "account",
	}
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
	invokeMultiAgentRotateKey()
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
		nil, func(accountAuthKey [32]byte, accountSeeds []string, tx *models.Transaction) error {
			originSeed1, err := hex.DecodeString(accountSeeds[0])
			if err != nil {
				return err
			}

			originSeed2, err := hex.DecodeString(accountSeeds[1])
			if err != nil {
				return err
			}

			originPriv1 := ed25519.NewKeyFromSeed(originSeed1)
			originPriv2 := ed25519.NewKeyFromSeed(originSeed2)

			_, newSeeds := createAccountTx(2)
			waitForTxConfirmed()
			if err := tx.SetSequenceNumber(tx.SequenceNumber + 1).Error(); err != nil {
				return err
			}

			newSeed1, err := hex.DecodeString(newSeeds[0])
			if err != nil {
				return err
			}

			newSeed2, err := hex.DecodeString(newSeeds[1])
			if err != nil {
				return err
			}

			newPriv1 := ed25519.NewKeyFromSeed(newSeed1)
			newPriv2 := ed25519.NewKeyFromSeed(newSeed2)
			newPublicKeys := append(newPriv1.Public().(ed25519.PublicKey), newPriv2.Public().(ed25519.PublicKey)...)

			type SigningMsg struct {
				TypeOf struct {
					Address [32]byte
					Module  string
					Struct  string
				}
				RotationProofChallenge struct {
					Seq        uint64
					Address    [32]byte
					AuthKey    [32]byte
					PublicKeys []byte
				}
			}

			signingMsg, err := lcs.Marshal(SigningMsg{
				TypeOf: struct {
					Address [32]byte
					Module  string
					Struct  string
				}{
					Address: addr0x1,
					Module:  "account",
					Struct:  "RotationProofChallenge",
				},
				RotationProofChallenge: struct {
					Seq        uint64
					Address    [32]byte
					AuthKey    [32]byte
					PublicKeys []byte
				}{
					Seq:        uint64(0),
					Address:    accountAuthKey,
					AuthKey:    accountAuthKey,
					PublicKeys: append(newPublicKeys, 0x01),
				},
			})
			if err != nil {
				return err
			}

			originSig := ed25519.Sign(originPriv1, signingMsg)
			newSig := ed25519.Sign(newPriv1, signingMsg)
			return tx.SetPayload(models.ScriptPayload{
				Code:          tx.Payload.(models.ScriptPayload).Code,
				TypeArguments: nil,
				Arguments: []models.TransactionArgument{
					models.TxArgU8{U8: 1},
					models.TxArgU8Vector{
						Bytes: append(append(originPriv1.Public().(ed25519.PublicKey), originPriv2.Public().(ed25519.PublicKey)...), 0x01),
					},
					models.TxArgU8{U8: 1},
					models.TxArgU8Vector{
						Bytes: append(append(newPriv1.Public().(ed25519.PublicKey), newPriv2.Public().(ed25519.PublicKey)...), 0x01),
					},
					models.TxArgU8Vector{
						Bytes: append(originSig, []byte{0x80, 0x00, 0x00, 0x00}...),
					},
					models.TxArgU8Vector{
						Bytes: append(newSig, []byte{0x80, 0x00, 0x00, 0x00}...),
					},
				},
			}).Error()
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
		}, uint64(1))
	fmt.Println("=====")
	waitForTxConfirmed()
	transferTxWeightedMultiED25519()
	fmt.Println("=====")
	waitForTxConfirmed()
	usdtCoinAddr, _ := models.HexToAccountAddress("0x498d8926f16eb9ca90cab1b3a26aa6f97a080b3fcbe6e83ae150b7243a00fb68")
	invokeMultiAgentScriptPayload("multi_agent_coin_register", []models.TypeTag{
		models.TypeTagStruct{
			Address: usdtCoinAddr,
			Module:  "devnet_coins",
			Name:    "DevnetUSDT",
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
	newPub, newPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	waitForTxConfirmed()
	faucet(authKey, 5000000)
	waitForTxConfirmed()

	accountInfo, err := aptosClient.GetAccount(ctx, address)
	if err != nil {
		panic(err)
	}

	type SigningMsg struct {
		TypeOf struct {
			Address [32]byte
			Module  string
			Struct  string
		}
		RotationProofChallenge struct {
			Seq        uint64
			Address    [32]byte
			AuthKey    [32]byte
			PublicKeys []byte
		}
	}

	seq, err := strconv.Atoi(accountInfo.SequenceNumber)
	if err != nil {
		panic(err)
	}

	signingMsg, err := lcs.Marshal(SigningMsg{
		TypeOf: struct {
			Address [32]byte
			Module  string
			Struct  string
		}{
			Address: addr0x1,
			Module:  "account",
			Struct:  "RotationProofChallenge",
		},
		RotationProofChallenge: struct {
			Seq        uint64
			Address    [32]byte
			AuthKey    [32]byte
			PublicKeys []byte
		}{
			Seq:        uint64(seq),
			Address:    authKey,
			AuthKey:    authKey,
			PublicKeys: newPub[:],
		},
	})
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(models.EntryFunctionPayload{
			Module:   accountModule,
			Function: "rotate_authentication_key",
			Arguments: []interface{}{
				uint8(0),
				[]byte(priv.Public().(ed25519.PublicKey)),
				uint8(0),
				[]byte(newPriv.Public().(ed25519.PublicKey)),
				ed25519.Sign(priv, signingMsg),
				ed25519.Sign(newPriv, signingMsg),
			},
		},
		).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: priv.Public().(ed25519.PublicKey),
		Signature: signature,
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	faucet(authKey, 5000000)
	waitForTxConfirmed()

	accountInfo, err := aptosClient.GetAccount(ctx, address)
	if err != nil {
		panic(err)
	}

	addr, _ := models.HexToAccountAddress(faucetAdminAddress)

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(getTransferPayload(addr, 1)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
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
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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

	accountInfo, err := aptosClient.GetAccount(ctx, faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}

	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(models.EntryFunctionPayload{
			Module:    aptosAccountModule,
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

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: faucetAdminPriv.Public().(ed25519.PublicKey),
		Signature: signature,
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	accountInfo, err := aptosClient.GetAccount(ctx, faucetAdminAddress)
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
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	signer := models.NewSingleSigner(priv)

	err = signer.Sign(&tx).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	senderInfo, err := aptosClient.GetAccount(ctx, sender)
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
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
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
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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

func invokeMultiAgentRotateKey() {
	authKey, seeds := createAccountTx(2)
	waitForTxConfirmed()
	faucet(authKey, 5000000)
	waitForTxConfirmed()
	originSeed1, err := hex.DecodeString(seeds[0])
	if err != nil {
		panic(err)
	}

	originSeed2, err := hex.DecodeString(seeds[1])
	if err != nil {
		panic(err)
	}

	originPriv1 := ed25519.NewKeyFromSeed(originSeed1)
	originPriv2 := ed25519.NewKeyFromSeed(originSeed2)

	_, newSeeds := createAccountTx(2)
	waitForTxConfirmed()
	newSeed1, err := hex.DecodeString(newSeeds[0])
	if err != nil {
		panic(err)
	}

	newSeed2, err := hex.DecodeString(newSeeds[1])
	if err != nil {
		panic(err)
	}

	newPriv1 := ed25519.NewKeyFromSeed(newSeed1)
	newPriv2 := ed25519.NewKeyFromSeed(newSeed2)
	newPublicKeys := append(newPriv1.Public().(ed25519.PublicKey), newPriv2.Public().(ed25519.PublicKey)...)

	type SigningMsg struct {
		TypeOf struct {
			Address [32]byte
			Module  string
			Struct  string
		}
		RotationProofChallenge struct {
			Seq        uint64
			Address    [32]byte
			AuthKey    [32]byte
			PublicKeys []byte
		}
	}

	signingMsg, err := lcs.Marshal(SigningMsg{
		TypeOf: struct {
			Address [32]byte
			Module  string
			Struct  string
		}{
			Address: addr0x1,
			Module:  "account",
			Struct:  "RotationProofChallenge",
		},
		RotationProofChallenge: struct {
			Seq        uint64
			Address    [32]byte
			AuthKey    [32]byte
			PublicKeys []byte
		}{
			Seq:        uint64(0),
			Address:    authKey,
			AuthKey:    authKey,
			PublicKeys: append(newPublicKeys, 0x01),
		},
	})
	if err != nil {
		panic(err)
	}

	originSig := ed25519.Sign(originPriv1, signingMsg)
	newSig := ed25519.Sign(newPriv1, signingMsg)
	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(hex.EncodeToString(authKey[:])).
		SetPayload(models.EntryFunctionPayload{
			Module:   accountModule,
			Function: "rotate_authentication_key",
			Arguments: []interface{}{
				uint8(1),
				[]byte(append(append(originPriv1.Public().(ed25519.PublicKey), originPriv2.Public().(ed25519.PublicKey)...), 0x01)),
				uint8(1),
				[]byte(append(append(newPriv1.Public().(ed25519.PublicKey), newPriv2.Public().(ed25519.PublicKey)...), 0x01)),
				append(originSig, []byte{0x80, 0x00, 0x00, 0x00}...),
				append(newSig, []byte{0x80, 0x00, 0x00, 0x00}...),
			},
		},
		).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(uint64(0)).
		Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(originPriv2, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorMultiEd25519{
		PublicKeys: []models.PublicKey{
			originPriv1.Public().(ed25519.PublicKey),
			originPriv2.Public().(ed25519.PublicKey),
		},
		Threshold:  1,
		Signatures: []models.Signature{signature},
		Bitmap:     [4]byte{0x40, 0, 0, 0},
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	faucet(authKey, 5000000)
	waitForTxConfirmed()

	accountInfo, err := aptosClient.GetAccount(ctx, address)
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
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		fmt.Printf("%+v\n", tx.UserTransaction)
		panic(err)
	}

	signature := ed25519.Sign(priv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: priv.Public().(ed25519.PublicKey),
		Signature: signature,
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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

func invokeMultiAgentScriptPayload(scriptName string, typeArgs []models.TypeTag, args []models.TransactionArgument, opts ...interface{}) {
	fileBytes, err := os.ReadFile(fmt.Sprintf("./examples/transactions/scripts/%s.mv", scriptName))
	if err != nil {
		panic(err)
	}

	authKey, seeds := createAccountTx(2)
	waitForTxConfirmed()

	var cb func(accountAuthKey [32]byte, accountSeeds []string, tx *models.Transaction) error
	for _, opt := range opts {
		switch value := opt.(type) {
		case uint64:
			faucet(authKey, value)
			waitForTxConfirmed()
		case func(accountAuthKey [32]byte, accountSeeds []string, tx *models.Transaction) error:
			cb = value
		}
	}

	sender := faucetAdminAddress
	senderInfo, err := aptosClient.GetAccount(ctx, sender)
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
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(senderInfo.SequenceNumber).
		SetSecondarySigners([]models.AccountAddress{authKey}).
		Error()
	if err != nil {
		panic(err)
	}

	if cb != nil {
		if err := cb(authKey, seeds, &tx); err != nil {
			panic(err)
		}
	}

	msgBytes, err := tx.GetSigningMessage()
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
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	accountInfo, err := aptosClient.GetAccount(ctx, faucetAdminAddress)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(faucetAdminAddress).
		SetPayload(models.EntryFunctionPayload{
			Module:    aptosAccountModule,
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

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		panic(err)
	}

	signature := ed25519.Sign(faucetAdminPriv, msgBytes)
	err = tx.SetAuthenticator(models.TransactionAuthenticatorEd25519{
		PublicKey: faucetAdminPriv.Public().(ed25519.PublicKey),
		Signature: signature,
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
	faucet(authKey, 5000000)
	waitForTxConfirmed()

	accountInfo, err := aptosClient.GetAccount(ctx, address)
	if err != nil {
		panic(err)
	}

	tx := models.Transaction{}
	err = tx.SetChainID(DevnetChainID).
		SetSender(address).
		SetPayload(getTransferPayload(faucetAdminAddr, 1)).
		SetExpirationTimestampSecs(uint64(time.Now().Add(10 * time.Minute).Unix())).
		SetGasUnitPrice(uint64(100)).
		SetMaxGasAmount(uint64(5000)).
		SetSequenceNumber(accountInfo.SequenceNumber).Error()
	if err != nil {
		panic(err)
	}

	msgBytes, err := tx.GetSigningMessage()
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
	}, true).Error()
	if err != nil {
		panic(err)
	}

	_, err = aptosClient.SimulateTransaction(ctx, tx.UserTransaction, false, false)
	if err != nil {
		panic(err)
	}

	rawTx, err := aptosClient.SubmitTransaction(ctx, tx.UserTransaction)
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
