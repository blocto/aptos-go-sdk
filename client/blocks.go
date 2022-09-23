package client

import (
	"context"
	"fmt"
	"net/http"
)

type Blocks interface {
	GetBlocksByHeight(ctx context.Context, height uint64, withTransactions bool, opts ...interface{}) (*Block, error)
	GetBlocksByVersion(ctx context.Context, version uint64, withTransactions bool, opts ...interface{}) (*Block, error)
}

type BlocksImpl struct {
	Base
}

type Block struct {
	BlockHeight    string            `json:"block_height"`
	BlockHash      string            `json:"block_hash"`
	BlockTimestamp string            `json:"block_timestamp"`
	FirstVersion   string            `json:"first_version"`
	LastVersion    string            `json:"last_version"`
	Transactions   []TransactionResp `json:"transactions"`
}

func (impl BlocksImpl) GetBlocksByHeight(ctx context.Context, height uint64, withTransactions bool, opts ...interface{}) (*Block, error) {
	var rspJSON Block
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/blocks/by_height/%d", height),
		nil, &rspJSON, map[string]interface{}{
			"with_transactions": withTransactions,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl BlocksImpl) GetBlocksByVersion(ctx context.Context, version uint64, withTransactions bool, opts ...interface{}) (*Block, error) {
	var rspJSON Block
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/blocks/by_version/%d", version),
		nil, &rspJSON, map[string]interface{}{
			"with_transactions": withTransactions,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
