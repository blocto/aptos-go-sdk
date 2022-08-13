package client

import (
	"fmt"
	"net/http"
)

type State interface {
	GetTableItemByHandleAndKey(handle string, req TableItemReq, opts ...interface{}) (*TableItemValue, error)
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

func (impl StateImpl) GetTableItemByHandleAndKey(handle string, req TableItemReq, opts ...interface{}) (*TableItemValue, error) {
	var rspJSON TableItemValue
	err := Request(http.MethodPost,
		impl.Base.Endpoint()+fmt.Sprintf("/tables/%s/item", handle),
		req, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &rspJSON, nil
}
