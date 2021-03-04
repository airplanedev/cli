// Package api implements Airplane HTTP API client.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const (
	// Host is the default HTTP endpoint.
	Host = "https://api.airplane.dev"
)

// Client implements Airplane client.
//
// The zero-value is ready for use.
type Client struct {
	// Host is the HTTP endpoint to use.
	//
	// If empty, it uses the global `api.Host`.
	Host string
}

// CreateTask creates a task with the given request.
func (c Client) CreateTask(ctx context.Context, req CreateTaskRequest) (res CreateTaskResponse, err error) {
	err = c.do(ctx, "POST", "/tasks/create", req, &res)
	return
}

// Do sends a request with `method`, `path`, `payload` and `reply`.
func (c Client) do(ctx context.Context, method, path string, payload, reply interface{}) error {
	var url = c.host() + path
	var body io.Reader

	// TODO(amir): validate before sending?
	//
	// maybe `if v, ok := payload.(validator); ok { v.validate() }`
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return errors.Wrap(err, "api: marshal payload")
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errors.Wrap(err, "api: new request")
	}

	resp, err := http.DefaultClient.Do(req)

	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}

	if err != nil {
		return errors.Wrapf(err, "api: %s %s", method, url)
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		return errors.Errorf("api: %s %s - %s", method, url, resp.Status)
	}

	if reply != nil {
		if err := json.NewDecoder(resp.Body).Decode(reply); err != nil {
			return errors.Wrapf(err, "api: %s %s - decoding json", method, url)
		}
	}

	return nil
}

// Host returns the configured endpoint.
func (c Client) host() string {
	if c.Host != "" {
		return c.Host
	}
	return Host
}
