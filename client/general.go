package client

import (
	"context"
	"net/http"
)

type General interface {
	LedgerInformation(ctx context.Context, opts ...interface{}) (*LedgerInfo, error)
	CheckBasicNodeHealth(ctx context.Context, durationSecs uint32, opts ...interface{}) (*HealthInfo, error)
}

type GeneralImp struct {
	Base
}

type LedgerInfo struct {
	ChainID             uint64 `json:"chain_id"`
	Epoch               string `json:"epoch"`
	LedgerVersion       string `json:"ledger_version"`
	LedgerTimestamp     string `json:"ledger_timestamp"`
	OldestLedgerVersion string `json:"oldest_ledger_version"`
	OldestBlockHeight   string `json:"oldest_block_height"`
	BlockHeight         string `json:"block_height"`
	NodeRole            string `json:"node_role"`
}

func (impl GeneralImp) LedgerInformation(ctx context.Context, opts ...interface{}) (*LedgerInfo, error) {
	var rspJSON LedgerInfo
	err := request(ctx, http.MethodGet, impl.Base.Endpoint()+"/v1", nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

type HealthInfo struct {
	Message string `json:"message"`
}

func (impl GeneralImp) CheckBasicNodeHealth(ctx context.Context, durationSecs uint32, opts ...interface{}) (*HealthInfo, error) {
	var rspJSON HealthInfo
	err := request(ctx, http.MethodGet, impl.Base.Endpoint()+"/v1/-/healthy",
		nil, &rspJSON, map[string]interface{}{
			"duration_secs": durationSecs,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
