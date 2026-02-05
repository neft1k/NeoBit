-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS clusters (
    id BIGSERIAL PRIMARY KEY,
    algorithm TEXT NOT NULL,
    k INT NOT NULL,
    centroid VECTOR(384) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    hn_id BIGINT,
    title TEXT,
    url TEXT,
    by TEXT,
    score INT,
    time TIMESTAMPTZ,
    text TEXT,
    embedding VECTOR(384) NOT NULL,
    cluster_id BIGINT REFERENCES clusters(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_documents_cluster_id ON documents (cluster_id);
CREATE INDEX IF NOT EXISTS idx_documents_embedding_hnsw ON documents USING hnsw (embedding vector_cosine_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_documents_embedding_hnsw;
DROP INDEX IF EXISTS idx_documents_cluster_id;

DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS clusters;

DROP EXTENSION IF EXISTS vector;
