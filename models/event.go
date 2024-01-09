package models

type Event struct {
	Version Uint64 `json:"version"`
	GUID    struct {
		CreationNumber Uint64 `json:"creation_number"`
		AccountAddress string `json:"account_address"`
	} `json:"guid"`
	SequenceNumber Uint64      `json:"sequence_number"`
	Type           string      `json:"type"`
	Data           interface{} `json:"data"`
}

type TokenDepositEvent struct {
	ID     TokenID `json:"id"`
	Amount string  `json:"amount"`
}
