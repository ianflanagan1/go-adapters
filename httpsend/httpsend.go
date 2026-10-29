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

// ErrBodyTooLarge is returned when a response body is longer than the
// client's limit.
var ErrBodyTooLarge = errors.New("response body too large")

// Client sends requests with HTTP, reading at most maxBody bytes of each
// response body.
type Client struct {
	http    *http.Client
	maxBody int64
}

// New returns a Client that sends requests with client and reads at most
// maxBody bytes of each response body. It panics if maxBody is below 1,
// which would reject every response with a body.
func New(client *http.Client, maxBody int64) Client {
	if maxBody < 1 {
		panic(fmt.Sprintf("httpsend.New: maxBody is %d; it must be at least 1", maxBody))
	}
	return Client{http: client, maxBody: maxBody}
}

// Send sends req and returns the response.
func (c Client) Send(ctx context.Context, req httpmsg.Request) (httpmsg.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("building the request: %w", err)
	}
	httpReq.Header = http.Header(req.Header).Clone()
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("sending %s %s: %w", req.Method, req.URL, err)
	}
	// The body is read in full or abandoned on error, so a failure to close
	// it tells the caller nothing.
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBody+1))
	if err != nil {
		return httpmsg.Response{}, fmt.Errorf("reading the response to %s %s: %w", req.Method, req.URL, err)
	}
	if int64(len(body)) > c.maxBody {
		return httpmsg.Response{}, fmt.Errorf("%s %s: %w", req.Method, req.URL, ErrBodyTooLarge)
	}
	return httpmsg.Response{Status: resp.StatusCode, Header: resp.Header, Body: body}, nil
}
