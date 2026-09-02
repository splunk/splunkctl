package override

import "testing"

func TestKvCollectionPath_WithApp(t *testing.T) {
	got := kvCollectionPath("myapp")
	want := "/servicesNS/nobody/myapp/storage/collections/config/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestKvCollectionPath_NoApp(t *testing.T) {
	got := kvCollectionPath("")
	want := "/servicesNS/nobody/search/storage/collections/config/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestKvDataPath_WithApp(t *testing.T) {
	got := kvDataPath("myapp", "mycoll")
	want := "/servicesNS/nobody/myapp/storage/collections/data/mycoll"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestKvDataPath_NoApp(t *testing.T) {
	got := kvDataPath("", "mycoll")
	want := "/servicesNS/nobody/search/storage/collections/data/mycoll"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEncodeParams_SinglePair(t *testing.T) {
	got := encodeParams(map[string]string{"key": "value"})
	if got != "key=value" {
		t.Errorf("got %q, want %q", got, "key=value")
	}
}

func TestEncodeParams_Empty(t *testing.T) {
	got := encodeParams(map[string]string{})
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestKvAppFromParams_FlagOnly(t *testing.T) {
	app, extra := kvAppFromParams("flagapp", map[string]string{"other": "val"})
	if app != "flagapp" {
		t.Errorf("got app=%q, want flagapp", app)
	}
	if _, ok := extra["other"]; !ok {
		t.Error("extra should be unchanged")
	}
}

func TestKvAppFromParams_ParamOnly(t *testing.T) {
	app, extra := kvAppFromParams("", map[string]string{"app": "paramapp", "other": "val"})
	if app != "paramapp" {
		t.Errorf("got app=%q, want paramapp", app)
	}
	if _, ok := extra["app"]; ok {
		t.Error("app key should be removed from extra")
	}
}

func TestKvAppFromParams_FlagWinsOverParam(t *testing.T) {
	app, extra := kvAppFromParams("flagapp", map[string]string{"app": "paramapp"})
	if app != "flagapp" {
		t.Errorf("got app=%q, want flagapp (flag should win)", app)
	}
	if _, ok := extra["app"]; ok {
		t.Error("app key should be removed from extra even when flag wins")
	}
}

func TestKvAppFromParams_NeitherSet(t *testing.T) {
	app, extra := kvAppFromParams("", map[string]string{"other": "val"})
	if app != "" {
		t.Errorf("got app=%q, want empty string", app)
	}
	if _, ok := extra["other"]; !ok {
		t.Error("extra should be unchanged")
	}
}

func TestKvParseEntries_SingleEntry(t *testing.T) {
	body := []byte(`{"entry":[{"name":"mycoll","content":{"count":42}}]}`)
	results, err := kvParseEntries(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "mycoll" {
		t.Errorf("expected name=mycoll, got %v", results[0]["name"])
	}
	if results[0]["count"] != float64(42) {
		t.Errorf("expected count=42, got %v", results[0]["count"])
	}
}

func TestKvParseEntries_MultipleEntries(t *testing.T) {
	body := []byte(`{"entry":[{"name":"a","content":{}},{"name":"b","content":{}}]}`)
	results, err := kvParseEntries(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestKvParseEntries_InvalidJSON(t *testing.T) {
	_, err := kvParseEntries([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestKvParseEntries_EmptyFeed(t *testing.T) {
	body := []byte(`{"entry":[]}`)
	results, err := kvParseEntries(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
