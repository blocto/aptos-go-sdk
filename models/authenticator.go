package models

import (
	"crypto/ed25519"
	"fmt"

	"github.com/the729/lcs"

	"github.com/portto/aptos-go-sdk/crypto"
)

type PrivateKey = ed25519.PrivateKey

type Signer interface {
	Sign(tx *Transaction) *Transaction
}

type SingleSigner struct {
	PrivateKey
	PublicKey
	AccountAddress
}

func NewSingleSigner(priv PrivateKey) SingleSigner {
	pub := priv.Public().(PublicKey)
	addr := crypto.SingleSignerAuthKey(pub)

	return SingleSigner{
		PrivateKey:     priv,
		PublicKey:      pub,
		AccountAddress: addr,
	}
}

func (s *SingleSigner) Sign(tx *Transaction) *Transaction {
	if tx.hasError() {
		return tx
	}

	msgBytes, err := tx.GetSigningMessage()
	if err != nil {
		tx.err = fmt.Errorf("GetSigningMessage error: %v", err)
		return tx
	}

	signature := ed25519.Sign(s.PrivateKey, msgBytes)

	return tx.SetAuthenticator(TransactionAuthenticatorEd25519{
		PublicKey: s.PublicKey,
		Signature: signature,
	})
}

type PublicKey = ed25519.PublicKey

type Signature []byte

type AccountAuthenticator interface{}

var _ = lcs.RegisterEnum(
	(*AccountAuthenticator)(nil),
	AccountAuthenticatorEd25519{},
	AccountAuthenticatorMultiEd25519{},
)

type AccountAuthenticatorEd25519 struct {
	PublicKey
	Signature
}

type AccountAuthenticatorMultiEd25519 struct {
	PublicKeyBytes []byte      // for BCS serialization; p1_bytes | ... | pn_bytes | threshold
	SignatureBytes []byte      // for BCS serialization; s1_bytes | ... | sn_bytes | bitmap
	PublicKeys     []PublicKey `lcs:"-"`
	Threshold      uint8       `lcs:"-"`
	Signatures     []Signature `lcs:"-"`
	Bitmap         [4]byte     `lcs:"-"`
}

func (aa AccountAuthenticatorMultiEd25519) SetBytes() AccountAuthenticatorMultiEd25519 {
	aa.PublicKeyBytes = make([]byte, ed25519.PublicKeySize*len(aa.PublicKeys)+1)
	for i, publicKey := range aa.PublicKeys {
		copy(aa.PublicKeyBytes[i*ed25519.PublicKeySize:], publicKey)
	}
	aa.PublicKeyBytes[len(aa.PublicKeys)*ed25519.PublicKeySize] = aa.Threshold

	aa.SignatureBytes = make([]byte, ed25519.SignatureSize*len(aa.Signatures)+4)
	for i, signature := range aa.Signatures {
		copy(aa.SignatureBytes[i*ed25519.SignatureSize:], signature)
	}
	copy(aa.SignatureBytes[len(aa.Signatures)*ed25519.SignatureSize:], aa.Bitmap[:])

	return aa
}

type TransactionAuthenticator interface{}

var _ = lcs.RegisterEnum(
	(*TransactionAuthenticator)(nil),
	TransactionAuthenticatorEd25519{},
	TransactionAuthenticatorMultiEd25519{},
	TransactionAuthenticatorMultiAgent{},
)

type TransactionAuthenticatorEd25519 struct {
	PublicKey
	Signature
}

type TransactionAuthenticatorMultiEd25519 struct {
	PublicKeyBytes []byte      // for BCS serialization; p1_bytes | ... | pn_bytes | threshold
	SignatureBytes []byte      // for BCS serialization; s1_bytes | ... | sn_bytes | bitmap
	PublicKeys     []PublicKey `lcs:"-"`
	Threshold      uint8       `lcs:"-"`
	Signatures     []Signature `lcs:"-"`
	Bitmap         [4]byte     `lcs:"-"`
}

func (txAuth TransactionAuthenticatorMultiEd25519) SetBytes() TransactionAuthenticatorMultiEd25519 {
	return TransactionAuthenticatorMultiEd25519(AccountAuthenticatorMultiEd25519(txAuth).SetBytes())
}

type TransactionAuthenticatorMultiAgent struct {
	Sender                   AccountAuthenticator
	SecondarySignerAddresses []AccountAddress
	SecondarySigners         []AccountAuthenticator
}
