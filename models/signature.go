package models

type Signature struct {
	Type string `json:"type"`
	ED25519Signature
	MultiED25519Signature
	MultiAgentSignature
}

type ED25519Signature struct {
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

type MultiED25519Signature struct {
	PublicKeys []string `json:"public_keys"`
	Signatures []string `json:"signatures"`
	Threshold  int      `json:"threshold"`
	Bitmap     string   `json:"bitmap"`
}

type MultiAgentSignature struct {
	Sender struct {
		Type string `json:"type"`
		ED25519Signature
		MultiED25519Signature
	} `json:"sender"`
	SecondarySignerAddresses []string `json:"secondary_signer_addresses"`
	SecondarySigners         []struct {
		Type string `json:"type"`
		ED25519Signature
		MultiED25519Signature
	} `json:"secondary_signers"`
}
