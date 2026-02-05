package document

import "time"

type CreateDocumentRequest struct {
	HNID      int64     `json:"hn_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	By        string    `json:"by"`
	Score     int       `json:"score"`
	Time      string    `json:"time"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
}

type CreateDocumentResponse struct {
	ID int64 `json:"id"`
}

type DocumentResponse struct {
	ID        int64     `json:"id"`
	HNID      int64     `json:"hn_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	By        string    `json:"by"`
	Score     int       `json:"score"`
	Time      time.Time `json:"time"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
	ClusterID *int64    `json:"cluster_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
