package cluster

import (
	"net/http"

	"NeoBIT/internal/logger"
	cluster_model "NeoBIT/internal/models/cluster"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	res, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		h.log.Error(r.Context(), "cluster list failed", logger.FieldAny("error", err))
		writeError(w, http.StatusInternalServerError, "failed to list clusters")
		return
	}
	writeJSON(w, http.StatusOK, toClusterResponses(res))
}

func toClusterResponses(clusters []cluster_model.Cluster) []cluster_model.ClusterResponse {
	out := make([]cluster_model.ClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		out = append(out, cluster_model.ClusterResponse(cluster))
	}
	return out
}
