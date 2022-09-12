package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTransactionByHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resp := []byte(mockTx)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			n, err := w.Write(resp)
			assert.NoError(t, err)
			assert.Equal(t, len(resp), n)
		}))
		impl := NewAptosClient(srv.URL)
		tx, err := impl.GetTransactionByHash(mockTxHash)
		assert.NoError(t, err)
		assert.Equal(t, "0x"+mockTxHash, tx.Hash)
	})
}

func TestWaitForTransaction(t *testing.T) {
	t.Run("Retry", func(t *testing.T) {
		resp := []byte(mockTx)
		var count int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if count == 0 {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, err := w.Write([]byte("bad request"))
				assert.NoError(t, err)
				count += 1
			} else {
				n, err := w.Write(resp)
				assert.NoError(t, err)
				assert.Equal(t, len(resp), n)
			}
		}))
		impl := NewAptosClient(srv.URL)
		err := impl.WaitForTransaction(mockTxHash)
		assert.NoError(t, err)
	})
}
