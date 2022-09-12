package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFund(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resp := []byte(fmt.Sprintf(`["%s"]`, mockTxHash))
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			n, err := w.Write(resp)
			assert.NoError(t, err)
			assert.Equal(t, len(resp), n)
		}))
		mockClient := MockAptosClient{}
		mockClient.On("WaitForTransaction", mockTxHash).Return(nil).Once()
		fc := NewFaucetClient(srv.URL, &mockClient)
		err := fc.FundAccount(mockCTX, mockAddress, 1)
		assert.NoError(t, err)
	})

	t.Run("Mint error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("bad request"))
		}))
		mockClient := MockAptosClient{}
		fc := NewFaucetClient(srv.URL, &mockClient)
		err := fc.FundAccount(mockCTX, mockAddress, 1)
		assert.EqualError(t, err, fmt.Sprintf("response(%d): %s", http.StatusServiceUnavailable, "bad request"))
	})
}
