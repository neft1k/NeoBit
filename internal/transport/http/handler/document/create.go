package document

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
)

func toDocumentModel(r document.CreateDocumentRequest) (document.Document, error) {
	if len(r.Embedding) == 0 {
		return document.Document{}, fmt.Errorf("embedding is required")
	}
	var ts time.Time
	if r.Time != "" {
		parsed, err := time.Parse(time.RFC3339, r.Time)
		if err != nil {
			return document.Document{}, fmt.Errorf("invalid time format")
		}
		ts = parsed
	} else {
		ts = time.Now()
	}

	return document.Document{
		HNID:      r.HNID,
		Title:     r.Title,
		URL:       r.URL,
		By:        r.By,
		Score:     r.Score,
		Time:      ts,
		Text:      r.Text,
		Embedding: r.Embedding,
	}, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req document.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn(r.Context(), "document create: invalid json", logger.FieldAny("error", err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	doc, err := toDocumentModel(req)
	if err != nil {
		h.log.Warn(r.Context(), "document create: invalid payload", logger.FieldAny("error", err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := h.svc.Create(r.Context(), doc)
	if err != nil {
		h.log.Error(r.Context(), "document create failed", logger.FieldAny("error", err))
		writeError(w, http.StatusInternalServerError, "failed to create document")
		return
	}
	writeJSON(w, http.StatusCreated, document.CreateDocumentResponse{ID: id})
}
