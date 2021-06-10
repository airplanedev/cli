package token

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Run("receive a token", func(t *testing.T) {
		var ctx = context.Background()
		var assert = require.New(t)

		srv, err := NewServer(ctx, "https://fake.airplane.so/cli/success")
		assert.NoError(err)

		send(t, srv.URL(), "token")

		assert.Equal("token", <-srv.Token())
		assert.NoError(srv.Close())
	})
}

func send(t testing.TB, url, token string) {
	t.Helper()

	req, err := http.NewRequest("GET", url+"?token="+token, nil)
	if err != nil {
		t.Fatalf("new request: %s", err)
	}

	var client = &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %s", err)
	}
	resp.Body.Close()
}
