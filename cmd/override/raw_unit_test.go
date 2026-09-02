package override

import "testing"

func TestNormalizeRawPath_StripServicesPrefix(t *testing.T) {
	got := normalizeRawPath("/services/data/indexes/")
	if got != "/data/indexes/" {
		t.Errorf("got %q, want %q", got, "/data/indexes/")
	}
}

func TestNormalizeRawPath_AlreadyClean(t *testing.T) {
	got := normalizeRawPath("/data/indexes/")
	if got != "/data/indexes/" {
		t.Errorf("got %q, want %q", got, "/data/indexes/")
	}
}

func TestNormalizeRawPath_NoLeadingSlash(t *testing.T) {
	got := normalizeRawPath("data/indexes/")
	if got != "/data/indexes/" {
		t.Errorf("got %q, want %q", got, "/data/indexes/")
	}
}
