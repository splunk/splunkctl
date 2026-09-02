package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestClusterSearchHeads_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/cluster/master/control/default/remove_search_heads" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		guids := r.FormValue("guids")
		if guids == "" {
			t.Error("guids param required for cluster-search-heads remove")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/cluster/master/control/default/remove_search_heads",
		Params: map[string]string{"guids": "guid-1234,guid-5678"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestClusterSearchHeads_Remove_SingleGUID(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/cluster/master/control/default/remove_search_heads" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("guids") != "single-guid-abcd" {
			t.Errorf("expected guids=single-guid-abcd, got %q", r.FormValue("guids"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/cluster/master/control/default/remove_search_heads",
		Params: map[string]string{"guids": "single-guid-abcd"},
	})
	if err != nil {
		t.Fatal(err)
	}
}
