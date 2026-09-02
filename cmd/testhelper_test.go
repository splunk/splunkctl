package cmd_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/config"
)

func newTLSTestServer(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	cfg := &config.Config{Host: srv.URL, Token: "test-token", Insecure: true}
	return client.New(cfg), srv
}
