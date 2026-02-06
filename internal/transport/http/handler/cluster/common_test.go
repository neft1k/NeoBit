package cluster

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseLimitOffset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/clusters?limit=7&offset=3", nil)
	limit, offset := parseLimitOffset(req)
	if limit != 7 || offset != 3 {
		t.Fatalf("expected limit=7 offset=3, got %d %d", limit, offset)
	}
}
