package document

import (
	"net/http"
	"strconv"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListByCluster(w http.ResponseWriter, r *http.Request) {
	clusterIDStr := chi.URLParam(r, "id")
	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil {
		h.log.Warn(r.Context(), "document list by cluster: invalid id", logger.FieldAny("error", err))
		writeError(w, http.StatusBadRequest, "invalid cluster id")
		return
	}
	limit, offset := parseLimitOffset(r)
	res, err := h.svc.ListByCluster(r.Context(), clusterID, limit, offset)
	if err != nil {
		h.log.Error(r.Context(), "document list by cluster failed", logger.FieldAny("error", err))
		writeError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	writeJSON(w, http.StatusOK, toDocumentResponses(res))
}

func toDocumentResponses(docs []document.Document) []document.DocumentResponse {
	out := make([]document.DocumentResponse, 0, len(docs))
	for _, doc := range docs {
		out = append(out, toDocumentResponse(doc))
	}
	return out
}
