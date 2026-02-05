package cluster

import "time"

type ClusterResponse struct {
	ID        int64     `json:"id"`
	Algorithm string    `json:"algorithm"`
	K         int       `json:"k"`
	Centroid  []float32 `json:"centroid"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
