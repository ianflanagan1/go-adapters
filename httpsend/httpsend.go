// Package httpsend sends httpmsg requests over HTTP. A response with any
// status is a Response, not an error; deciding what a status means is the
// caller's job.
package httpsend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ianflanagan1/go-ports/httpmsg"
)

// ErrBodyTooLarge is returned when a response body is longer than MaxBody.
var ErrBodyTooLarge = errors.New("response body too large")

// Client sends requests with HTTP, reading at most MaxBody bytes of each
// response body.
type Client struct {
	HTTP    *http.Client
	MaxBody int64
}

// Send sends req and returns the response.
func (c Client) Send(ctx context.Context, req httpmsg.Request) (httpmsg.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("building the request: %w", err)
	}
	httpReq.Header = http.Header(req.Header).Clone()
	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("sending %s %s: %w", req.Method, req.URL, err)
	}
	// The body is read in full or abandoned on error, so a failure to close
	// it tells the caller nothing.
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.MaxBody+1))
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("reading the response to %s %s: %w", req.Method, req.URL, err)
	}
	if int64(len(body)) > c.MaxBody {
		return httpmsg.Response{}, fmt.Errorf("%s %s: %w", req.Method, req.URL, ErrBodyTooLarge)
	}
	return httpmsg.Response{Status: resp.StatusCode, Header: resp.Header, Body: body}, nil
}
