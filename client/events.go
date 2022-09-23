package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/portto/aptos-go-sdk/models"
)

type Events interface {
	GetEventsByEventKey(ctx context.Context, key string, opts ...interface{}) ([]models.Event, error)
	GetEventsByEventHandle(ctx context.Context, address, handleStruct, fieldName string, start, limit int, opts ...interface{}) ([]models.Event, error)
}

type EventsImpl struct {
	Base
}

func (impl EventsImpl) GetEventsByEventKey(ctx context.Context, key string, opts ...interface{}) ([]models.Event, error) {
	var rspJSON []models.Event
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/events/%s", key),
		nil, &rspJSON, nil, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl EventsImpl) GetEventsByEventHandle(ctx context.Context, address, handleStruct, fieldName string, start, limit int, opts ...interface{}) ([]models.Event, error) {
	var rspJSON []models.Event
	err := request(ctx, http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/v1/accounts/%s/events/%s/%s",
			address, handleStruct, fieldName),
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		}, requestOptions(opts...))
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}
