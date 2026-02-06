package document

import (
	"testing"
	"time"

	"NeoBIT/internal/models/document"
)

func TestToDocumentModel(t *testing.T) {
	_, err := toDocumentModel(document.CreateDocumentRequest{})
	if err == nil {
		t.Fatalf("expected error on empty embedding")
	}

	req := document.CreateDocumentRequest{
		HNID:      1,
		Title:     "t",
		By:        "u",
		Score:     1,
		Time:      time.Now().UTC().Format(time.RFC3339),
		Text:      "x",
		Embedding: []float32{0.1, 0.2},
	}
	got, err := toDocumentModel(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.HNID != req.HNID {
		t.Fatalf("HNID mismatch")
	}
}
