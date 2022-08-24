package models

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/the729/lcs"
)

type PublicKey = ed25519.PublicKey

type Signature []byte

func toHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

type AccountAuthenticator interface {
	ToJSONSigner() JSONSigner
}

var _ = lcs.RegisterEnum(
	(*AccountAuthenticator)(nil),
	AccountAuthenticatorEd25519{},
	AccountAuthenticatorMultiEd25519{},
)

type AccountAuthenticatorEd25519 struct {
	PublicKey
	Signature
}

func (aa AccountAuthenticatorEd25519) ToJSONSigner() JSONSigner {
	return JSONSigner{
		Type: "ed25519_signature",
		ED25519Signature: ED25519Signature{
			PublicKey: toHex(aa.PublicKey),
			Signature: toHex(aa.Signature),
		},
	}
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

func (aa AccountAuthenticatorMultiEd25519) ToJSONSigner() JSONSigner {
	json := JSONSigner{
		Type: "multi_ed25519_signature",
	}

	json.Threshold = aa.Threshold
	json.Bitmap = hex.EncodeToString(aa.Bitmap[:])

	for _, publicKey := range aa.PublicKeys {
		json.PublicKeys = append(json.PublicKeys, toHex(publicKey))
	}

	for _, signature := range aa.Signatures {
		json.Signatures = append(json.Signatures, toHex(signature))
	}

	return json
}

type TransactionAuthenticator interface {
	ToJSONSignature() *JSONSignature
}

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

func (txAuth TransactionAuthenticatorEd25519) ToJSONSignature() *JSONSignature {
	return &JSONSignature{
		Type: "ed25519_signature",
		ED25519Signature: ED25519Signature{
			PublicKey: toHex(txAuth.PublicKey),
			Signature: toHex(txAuth.Signature),
		},
	}
}

func (txAuth TransactionAuthenticatorMultiEd25519) ToJSONSignature() *JSONSignature {
	json := &JSONSignature{
		Type: "multi_ed25519_signature",
	}

	json.Threshold = txAuth.Threshold
	json.Bitmap = hex.EncodeToString(txAuth.Bitmap[:])

	for _, publicKey := range txAuth.PublicKeys {
		json.PublicKeys = append(json.PublicKeys, toHex(publicKey))
	}

	for _, signature := range txAuth.Signatures {
		json.Signatures = append(json.Signatures, toHex(signature))
	}

	return json
}

func (txAuth TransactionAuthenticatorMultiAgent) ToJSONSignature() *JSONSignature {
	json := &JSONSignature{
		Type: "multi_agent_signature",
	}

	json.Sender = txAuth.Sender.ToJSONSigner()

	for _, addr := range txAuth.SecondarySignerAddresses {
		json.SecondarySignerAddresses = append(json.SecondarySignerAddresses, addr.PrefixZeroTrimmedHex())
	}

	for _, signer := range txAuth.SecondarySigners {
		json.SecondarySigners = append(json.SecondarySigners, signer.ToJSONSigner())
	}

	return json
}
