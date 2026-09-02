package override

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/config"
)

func newSearchClient(t *testing.T, handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	return c, srv
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "")
	cmd.SetContext(context.Background())
	return cmd
}

func pollJobJSON(isDone bool, dispatchState string) string {
	return fmt.Sprintf(`{"entry":[{"content":{"isDone":%t,"dispatchState":%q}}]}`, isDone, dispatchState)
}

// fetchResults tests

func TestFetchResults_HappyPath(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[{"_raw":"event one"},{"_raw":"event two"}]}`)
	})
	defer srv.Close()

	err := fetchResults(newSearchCmd(), c, "mysid", 100, "json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFetchResults_ClampsMaxoutTooLow(t *testing.T) {
	var gotCount string
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCount = r.URL.Query().Get("count")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[]}`)
	})
	defer srv.Close()

	_ = fetchResults(newSearchCmd(), c, "mysid", 0, "")
	if gotCount != "100" {
		t.Errorf("expected count=100 when maxout=0, got count=%q", gotCount)
	}
}

func TestFetchResults_ClampsMaxoutTooHigh(t *testing.T) {
	var gotCount string
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCount = r.URL.Query().Get("count")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[]}`)
	})
	defer srv.Close()

	_ = fetchResults(newSearchCmd(), c, "mysid", 99999, "")
	if gotCount != "100" {
		t.Errorf("expected count=100 when maxout=99999, got count=%q", gotCount)
	}
}

func TestFetchResults_HTTPError(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"messages":[{"type":"ERROR","text":"server error"}]}`)
	})
	defer srv.Close()

	err := fetchResults(newSearchCmd(), c, "mysid", 100, "")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "results fetch failed") {
		t.Errorf("expected 'results fetch failed' in error, got %v", err)
	}
}

// pollJob tests

func TestPollJob_IsDoneSuccess(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, pollJobJSON(true, "DONE"))
	})
	defer srv.Close()

	err := pollJob(context.Background(), c, "mysid", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPollJob_FailedState(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, pollJobJSON(true, "FAILED"))
	})
	defer srv.Close()

	err := pollJob(context.Background(), c, "mysid", 0)
	if err == nil {
		t.Fatal("expected error for FAILED state")
	}
	if !strings.Contains(err.Error(), "failed with state") {
		t.Errorf("expected 'failed with state' in error, got %v", err)
	}
}

func TestPollJob_ErrorState(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, pollJobJSON(true, "ERROR"))
	})
	defer srv.Close()

	err := pollJob(context.Background(), c, "mysid", 0)
	if err == nil {
		t.Fatal("expected error for ERROR state")
	}
	if !strings.Contains(err.Error(), "failed with state") {
		t.Errorf("expected 'failed with state' in error, got %v", err)
	}
}

func TestPollJob_HTTPError(t *testing.T) {
	c, srv := newSearchClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal error")
	})
	defer srv.Close()

	err := pollJob(context.Background(), c, "mysid", 0)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "job status check failed") {
		t.Errorf("expected 'job status check failed' in error, got %v", err)
	}
}
