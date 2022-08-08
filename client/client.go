package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
}

func New(endpoint string) API {
	impl := &APIImpl{
		APIBase: APIBase{endpoint: endpoint},
	}

	impl.GeneralImp.Base = impl.APIBase
	impl.AccountsImpl.Base = impl.APIBase
	impl.StateImpl.Base = impl.APIBase
	impl.EventsImpl.Base = impl.APIBase
	impl.TransactionsImpl.Base = impl.APIBase
	return impl
}

type APIImpl struct {
	APIBase

	GeneralImp
	TransactionsImpl
	AccountsImpl
	EventsImpl
	StateImpl
}

type APIBase struct {
	endpoint string
}

func (impl APIBase) Endpoint() string {
	return impl.endpoint
}

type Base interface {
	Endpoint() string
}

type API interface {
	General
	Transactions
	Accounts
	Events
	State
}

func Request(method, endpoint string, reqBody, resp interface{}, query map[string]interface{}) error {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	if req.URL != nil && query != nil {
		query := req.URL.Query()
		for k, v := range query {
			query.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(context.Background())
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	rspBody, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("response(%d): %s", rsp.StatusCode, string(rspBody))
	}

	err = json.Unmarshal(rspBody, resp)
	if err != nil {
		return err
	}

	return nil
}
