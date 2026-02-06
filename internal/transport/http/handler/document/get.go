package document

import (
	"net/http"
	"strconv"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warn(r.Context(), "document get: invalid id", logger.FieldAny("error", err))
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	res, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.log.Warn(r.Context(), "document get: not found", logger.FieldAny("error", err))
		writeError(w, http.StatusNotFound, "document not found")
		return
	}
	writeJSON(w, http.StatusOK, toDocumentResponse(res))
}

func toDocumentResponse(doc document.Document) document.DocumentResponse {
	return document.DocumentResponse(doc)
}
