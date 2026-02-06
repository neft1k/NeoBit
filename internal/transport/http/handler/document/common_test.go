package document

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseLimitOffset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/documents?limit=5&offset=10", nil)
	limit, offset := parseLimitOffset(req)
	if limit != 5 || offset != 10 {
		t.Fatalf("expected limit=5 offset=10, got %d %d", limit, offset)
	}

	req = httptest.NewRequest(http.MethodGet, "/documents", nil)
	limit, offset = parseLimitOffset(req)
	if limit != 20 || offset != 0 {
		t.Fatalf("expected defaults 20/0, got %d %d", limit, offset)
	}
}
