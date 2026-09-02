package override

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalJSON_ValidJSON(t *testing.T) {
	var m map[string]string
	err := unmarshalJSON([]byte(`{"key":"value"}`), &m)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("expected key=value, got %q", m["key"])
	}
}

func TestUnmarshalJSON_InvalidJSON(t *testing.T) {
	var m map[string]string
	err := unmarshalJSON([]byte(`not json`), &m)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	var syntaxErr *json.SyntaxError
	if err == syntaxErr {
		t.Error("expected a JSON syntax error")
	}
}

func TestUrlValues_ConvertsMapsCorrectly(t *testing.T) {
	v := urlValues(map[string]string{"search": "index=main", "count": "10"})
	if v.Get("search") != "index=main" {
		t.Errorf("expected search=index=main, got %q", v.Get("search"))
	}
	if v.Get("count") != "10" {
		t.Errorf("expected count=10, got %q", v.Get("count"))
	}
}
