// Package api implements Airplane HTTP API client.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	// Endpoint is the default HTTP endpoint.
	Endpoint = "https://api.airplane.dev"
)

// Client implemnets Airplane client.
//
// The zero-value is ready for use.
type Client struct {
	// Endpoint is the HTTP endpoint to use.
	//
	// If empty, it uses the global `api.Endpoint`.
	Endpoint string
}

// CreateTask creates a task with the given request.
func (c Client) CreateTask(ctx context.Context, req CreateTaskRequest) (res CreateTaskResponse, err error) {
	err = c.do(ctx, "POST", "/tasks/create", req, &res)
	return
}

// Do sends a request with `method`, `path`, `payload` and `reply`.
func (c Client) do(ctx context.Context, method, path string, payload, reply interface{}) error {
	var url = c.endpoint() + path
	var body io.Reader

	// TODO(amir): validate before sending?
	//
	// maybe `if v, ok := payload.(validator); ok { v.validate() }`
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("api: marshal payload %T - %w", payload, err)
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("api: new request - %w", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}

	if err != nil {
		return fmt.Errorf("api: %s %s - %w", method, url, err)
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		return fmt.Errorf("api: %s %s - %s", method, url, resp.Status)
	}

	if reply != nil {
		if err := json.NewDecoder(resp.Body).Decode(reply); err != nil {
			return fmt.Errorf("api: %s %s - decoding json - %w", method, url, err)
		}
	}

	return nil
}

// Endpoint returns the configured endpoint.
func (c Client) endpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	return Endpoint
}
