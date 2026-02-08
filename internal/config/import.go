package config

import "time"

type ImportConfig struct {
	Enabled              bool
	DatasetURL           string
	LocalPath            string
	Limit                int
	ReadBatchSize        int
	WriteBatchSize       int
	ShutdownTimeout      time.Duration
	SkipIfDocumentsExist bool
}

func GetImportConfig() ImportConfig {
	return ImportConfig{
		Enabled:              getEnvBool("IMPORT_ENABLED", false),
		DatasetURL:           getEnv("IMPORT_DATASET_URL", "https://clickhouse-datasets.s3.amazonaws.com/hackernews-miniLM/hackernews_part_1_of_1.parquet"),
		LocalPath:            getEnv("IMPORT_LOCAL_PATH", "/tmp/hackernews.parquet"),
		Limit:                getEnvInt("IMPORT_LIMIT", 200000),
		ReadBatchSize:        getEnvInt("IMPORT_READ_BATCH_SIZE", 1000),
		WriteBatchSize:       getEnvInt("IMPORT_WRITE_BATCH_SIZE", 500),
		ShutdownTimeout:      time.Duration(getEnvInt("IMPORT_SHUTDOWN_TIMEOUT_SEC", 30)) * time.Second,
		SkipIfDocumentsExist: getEnvBool("IMPORT_SKIP_IF_DOCS_EXIST", true),
	}
}

func getEnvInt(key string, def int) int {
	v := getEnv(key, "")
	if v == "" {
		return def
	}

	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}
