package client

import (
	"context"
	"fmt"
	"net/http"
)

type State interface {
	GetTableItemByHandleAndKey(ctx context.Context, handle string, req TableItemReq, resp interface{}, opts ...interface{}) error
}

type StateImpl struct {
	Base
}

type TableItemReq struct {
	KeyType   string      `json:"key_type"`
	ValueType string      `json:"value_type"`
	Key       interface{} `json:"key"`
}

func (impl StateImpl) GetTableItemByHandleAndKey(ctx context.Context, handle string, req TableItemReq, resp interface{}, opts ...interface{}) error {
	err := request(ctx, http.MethodPost,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/tables/%s/item", handle),
		req, resp, nil, requestOptions(opts...))
	if err != nil {
		return err
	}

	return nil
}
