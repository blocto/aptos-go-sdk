package client

import (
	"context"
	"fmt"
	"net/http"
)

type State interface {
	GetTableItemByHandleAndKey(ctx context.Context, handle string, req TableItemReq, opts ...interface{}) (*TableItemValue, error)
}

type StateImpl struct {
	Base
}

type TableItemReq struct {
	KeyType   string `json:"key_type"`
	ValueType string `json:"value_type"`
	Key       string `json:"key"`
}

type TableItemValue struct {
}

func (impl StateImpl) GetTableItemByHandleAndKey(ctx context.Context, handle string, req TableItemReq, opts ...interface{}) (*TableItemValue, error) {
	var rspJSON TableItemValue
	err := request(ctx, http.MethodPost,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/tables/%s/item", handle),
		req, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
