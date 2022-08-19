package client

import (
	"fmt"
	"net/http"
)

type Blocks interface {
	GetBlocksByHeight(height uint64, withTransactions bool, opts ...interface{}) (*Block, error)
	GetBlocksByVersion(version uint64, withTransactions bool, opts ...interface{}) (*Block, error)
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

func (impl BlocksImpl) GetBlocksByHeight(height uint64, withTransactions bool, opts ...interface{}) (*Block, error) {
	var rspJSON Block
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/blocks/by_height/%d", height),
		nil, &rspJSON, map[string]interface{}{
			"with_transactions": withTransactions,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}

func (impl BlocksImpl) GetBlocksByVersion(version uint64, withTransactions bool, opts ...interface{}) (*Block, error) {
	var rspJSON Block
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/blocks/by_version/%d", version),
		nil, &rspJSON, map[string]interface{}{
			"with_transactions": withTransactions,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
