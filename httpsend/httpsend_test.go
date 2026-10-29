package httpsend_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ianflanagan1/go-adapters/httpsend"
	"github.com/ianflanagan1/go-ports/httpmsg"
)

func TestARequestRoundTrips(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Echo", r.Header.Get("X-Test"))
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write(append([]byte(r.Method+" "), body...))
	}))
	defer server.Close()

	client := httpsend.New(server.Client(), 1024)
	resp, err := client.Send(context.Background(), httpmsg.Request{
		Method: http.MethodPost, URL: server.URL, Header: map[string][]string{"X-Test": {"yes"}}, Body: []byte("hi"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != http.StatusTeapot || string(resp.Body) != "POST hi" || resp.Header["X-Echo"][0] != "yes" {
		t.Fatalf("got %+v", resp)
	}
}

func TestABodyOverTheLimitIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer server.Close()

	client := httpsend.New(server.Client(), 9)
	_, err := client.Send(context.Background(), httpmsg.Request{Method: http.MethodGet, URL: server.URL})
	if !errors.Is(err, httpsend.ErrBodyTooLarge) {
		t.Fatalf("got %v, want ErrBodyTooLarge", err)
	}
}

func TestAnUnreachableServerIsAnError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL
	server.Close()

	client := httpsend.New(http.DefaultClient, 1024)
	if _, err := client.Send(context.Background(), httpmsg.Request{Method: http.MethodGet, URL: url}); err == nil {
		t.Fatal("got no error")
	}
}

func TestABodyLimitBelowOneIsAProgrammingError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("got no panic")
		}
	}()
	httpsend.New(http.DefaultClient, 0)
}
