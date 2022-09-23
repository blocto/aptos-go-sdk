package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMsg := "succ"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, err := w.Write([]byte([]byte(fmt.Sprintf(`"%s"`, mockMsg))))
			assert.NoError(t, err)
		}))
		var resp string
		err := request(mockCTX, "GET", srv.URL, nil, &resp, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, mockMsg, resp)
	})

	t.Run("Timeout", func(t *testing.T) {
		mockMsg := "succ"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(1 * time.Second)
			_, err := w.Write([]byte([]byte(fmt.Sprintf(`"%s"`, mockMsg))))
			assert.NoError(t, err)
		}))
		var resp string
		ctx, cancel := context.WithTimeout(mockCTX, 10*time.Millisecond)
		defer cancel()
		err := request(ctx, "GET", srv.URL, nil, &resp, nil, nil)
		assert.Equal(t, true, errors.Is(err, context.DeadlineExceeded))
		assert.Equal(t, "", resp)
	})
}
