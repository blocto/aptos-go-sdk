package client

import (
	"net/http"
)

type General interface {
	LedgerInformation() (*LedgerInfo, error)
}

type GeneralImp struct {
	Base
}

type LedgerInfo struct {
	ChainID         uint64 `json:"chain_id"`
	LedgerVersion   string `json:"ledger_version"`
	LedgerTimestamp string `json:"ledger_timestamp"`
}

func (impl GeneralImp) LedgerInformation() (*LedgerInfo, error) {
	var rspJSON LedgerInfo
	err := Request(http.MethodGet, impl.Base.Endpoint(), nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
