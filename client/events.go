package client

import (
	"fmt"
	"net/http"
)

type Events interface {
	GetEventsByEventKey(key string) ([]Event, error)
	GetEventsByEventHandle(address, handleStruct, fieldName string, start, limit int) ([]Event, error)
}

type EventsImpl struct {
	Base
}

type Event struct {
	Key            string                 `json:"key"`
	SequenceNumber string                 `json:"sequence_number"`
	Type           string                 `json:"type"`
	Data           map[string]interface{} `json:"data"`
}

func (impl EventsImpl) GetEventsByEventKey(key string) ([]Event, error) {
	var rspJSON []Event
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/events/%s", key),
		nil, &rspJSON, nil)
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}

func (impl EventsImpl) GetEventsByEventHandle(address, handleStruct, fieldName string, start, limit int) ([]Event, error) {
	var rspJSON []Event
	err := Request(http.MethodGet,
		impl.Base.Endpoint()+fmt.Sprintf("/accounts/%s/events/%s/%s",
			address, handleStruct, fieldName),
		nil, &rspJSON, map[string]interface{}{
			"start": start,
			"limit": limit,
		})
	if err != nil {
		return nil, err
	}

	return rspJSON, nil
}
