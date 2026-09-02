package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestCredential_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/storage/passwords/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "myrealm:myuser:",
					"content": map[string]any{
						"username":       "myuser",
						"realm":          "myrealm",
						"clear_password": "********",
						"encr_password":  "some_encrypted_value",
					},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/storage/passwords/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(results))
	}
	if results[0]["name"] != "myrealm:myuser:" {
		t.Errorf("expected myrealm:myuser:, got %v", results[0]["name"])
	}
}

func TestCredential_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.FormValue("name") == "" || r.FormValue("password") == "" || r.FormValue("realm") == "" {
			t.Error("name, password, realm required for credential add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name":    r.FormValue("realm") + ":" + r.FormValue("name") + ":",
					"content": map[string]any{"username": r.FormValue("name"), "realm": r.FormValue("realm")},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/storage/passwords/",
		Params: map[string]string{"name": "myuser", "password": "s3cr3t", "realm": "myrealm"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "myrealm:myuser:" {
		t.Errorf("expected myrealm:myuser:, got %v", results)
	}
}

func TestCredential_Edit(t *testing.T) {
	credID := "myrealm:myuser:"
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "myrealm") {
			t.Errorf("expected realm in path, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("password") == "" {
			t.Error("password required for credential edit")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": credID, "content": map[string]any{"username": "myuser"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/storage/passwords/",
		ID:     credID,
		Params: map[string]string{"password": "newpassword"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCredential_Remove(t *testing.T) {
	credID := "myrealm:myuser:"
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "myrealm") || !strings.Contains(r.URL.Path, "myuser") {
			t.Errorf("expected realm:user in path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/storage/passwords/",
		ID:     credID,
	})
	if err != nil {
		t.Fatal(err)
	}
}
