package importer

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"NeoBIT/internal/config"
	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
)

type ImportService struct {
	docRepo DocumentRepository
	cfg     config.ImportConfig
	log     logger.Logger
}

func NewService(docRepo DocumentRepository, cfg config.ImportConfig, log logger.Logger) *ImportService {
	if log == nil {
		log = logger.Nop()
	}
	return &ImportService{
		docRepo: docRepo,
		cfg:     cfg,
		log:     log,
	}
}

func (s *ImportService) Run(ctx context.Context) error {
	if !s.cfg.Enabled {
		s.log.Info(ctx, "dataset import is disabled")
		return nil
	}
	if s.docRepo == nil {
		return fmt.Errorf("import service: document repo is nil")
	}
	if err := s.waitForDocumentsTable(ctx); err != nil {
		return err
	}

	if s.cfg.SkipIfDocumentsExist {
		count, err := s.docRepo.Count(ctx)
		if err != nil {
			return fmt.Errorf("import service: count documents: %w", err)
		}
		if count > 0 {
			s.log.Info(ctx, "dataset import skipped: documents already exist", logger.FieldAny("documents", count))
			return nil
		}
	}

	path, err := s.prepareDatasetFile(ctx)
	if err != nil {
		return err
	}

	inserted, scanned, err := s.importByPath(ctx, path)
	if err != nil {
		return err
	}

	s.log.Info(
		ctx,
		"dataset import finished",
		logger.FieldAny("rows_scanned", scanned),
		logger.FieldAny("rows_inserted", inserted),
	)
	return nil
}

func (s *ImportService) importByPath(ctx context.Context, path string) (inserted int64, scanned int64, err error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv":
		return s.importFromCSV(ctx, path)
	case ".parquet":
		return s.importFromParquet(ctx, path)
	default:
		return 0, 0, fmt.Errorf("import service: unsupported dataset extension: %s", filepath.Ext(path))
	}
}

type datasetRow struct {
	DocID     int32     `parquet:"name=doc_id, type=INT32"`
	Title     string    `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"`
	By        string    `parquet:"name=by, type=BYTE_ARRAY, convertedtype=UTF8"`
	PostScore int32     `parquet:"name=post_score, type=INT32"`
	Time      int32     `parquet:"name=time, type=INT32"`
	Text      string    `parquet:"name=text, type=BYTE_ARRAY, convertedtype=UTF8"`
	Vector    []float32 `parquet:"name=vector, type=LIST"`
}

func (s *ImportService) importFromParquet(ctx context.Context, path string) (inserted int64, scanned int64, err error) {
	fileReader, err := local.NewLocalFileReader(path)
	if err != nil {
		return 0, 0, fmt.Errorf("import service: open parquet file: %w", err)
	}
	defer fileReader.Close()

	parquetReader, err := reader.NewParquetReader(fileReader, new(datasetRow), 1)
	if err != nil {
		return 0, 0, fmt.Errorf("import service: create parquet reader: %w", err)
	}
	defer parquetReader.ReadStop()

	totalRows := int(parquetReader.GetNumRows())
	limit := s.cfg.Limit
	if limit <= 0 || limit > totalRows {
		limit = totalRows
	}
	if limit == 0 {
		return 0, 0, nil
	}

	readBatchSize := s.cfg.ReadBatchSize
	if readBatchSize <= 0 {
		readBatchSize = 1000
	}

	for scanned < int64(limit) {
		if err := ctx.Err(); err != nil {
			return inserted, scanned, err
		}

		left := limit - int(scanned)
		size := readBatchSize
		if size > left {
			size = left
		}

		rows := make([]datasetRow, size)
		if err := parquetReader.Read(&rows); err != nil {
			if err == io.EOF {
				break
			}
			return inserted, scanned, fmt.Errorf("import service: read parquet batch: %w", err)
		}
		scanned += int64(len(rows))

		docs := make([]document.Document, 0, len(rows))
		for _, row := range rows {
			if len(row.Vector) != 384 {
				continue
			}

			docs = append(docs, document.Document{
				HNID:      int64(row.DocID),
				Title:     row.Title,
				URL:       "",
				By:        row.By,
				Score:     int(row.PostScore),
				Time:      time.Unix(int64(row.Time), 0).UTC(),
				Text:      row.Text,
				Embedding: row.Vector,
			})
		}

		n, err := s.insertInBatches(ctx, docs)
		if err != nil {
			return inserted, scanned, err
		}
		inserted += n

		if scanned%10000 == 0 {
			s.log.Info(
				ctx,
				"dataset import progress",
				logger.FieldAny("rows_scanned", scanned),
				logger.FieldAny("rows_inserted", inserted),
				logger.FieldAny("limit", limit),
			)
		}
	}

	return inserted, scanned, nil
}

func (s *ImportService) importFromCSV(ctx context.Context, path string) (inserted int64, scanned int64, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("import service: open csv file: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(bufio.NewReader(file))
	r.FieldsPerRecord = -1
	r.ReuseRecord = true

	header, err := r.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("import service: read csv header: %w", err)
	}
	indexByColumn := make(map[string]int, len(header))
	for i, c := range header {
		indexByColumn[strings.ToLower(strings.TrimSpace(c))] = i
	}

	idxDocID, ok := indexByColumn["doc_id"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing doc_id column")
	}
	idxTitle, ok := indexByColumn["title"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing title column")
	}
	idxBy, ok := indexByColumn["by"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing by column")
	}
	idxScore, ok := indexByColumn["post_score"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing post_score column")
	}
	idxText, ok := indexByColumn["text"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing text column")
	}
	idxVector, ok := indexByColumn["vector"]
	if !ok {
		return 0, 0, fmt.Errorf("import service: csv missing vector column")
	}

	idxTime := -1
	if i, exists := indexByColumn["time_unix"]; exists {
		idxTime = i
	} else if i, exists := indexByColumn["time"]; exists {
		idxTime = i
	}
	if idxTime == -1 {
		return 0, 0, fmt.Errorf("import service: csv missing time column")
	}

	limit := s.cfg.Limit
	if limit <= 0 {
		limit = 200000
	}

	readBatchSize := s.cfg.ReadBatchSize
	if readBatchSize <= 0 {
		readBatchSize = 1000
	}

	docs := make([]document.Document, 0, readBatchSize)
	for scanned < int64(limit) {
		if err := ctx.Err(); err != nil {
			return inserted, scanned, err
		}

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return inserted, scanned, fmt.Errorf("import service: read csv record: %w", err)
		}
		scanned++

		docID, err := strconv.ParseInt(strings.TrimSpace(record[idxDocID]), 10, 64)
		if err != nil {
			continue
		}

		score := 0
		if v := strings.TrimSpace(record[idxScore]); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				score = parsed
			}
		}

		embedding, err := parseVector(record[idxVector])
		if err != nil || len(embedding) != 384 {
			continue
		}

		sec := int64(0)
		if v := strings.TrimSpace(record[idxTime]); v != "" {
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				sec = parsed
			}
		}
		ts := time.Unix(sec, 0).UTC()

		docs = append(docs, document.Document{
			HNID:      docID,
			Title:     record[idxTitle],
			URL:       "",
			By:        record[idxBy],
			Score:     score,
			Time:      ts,
			Text:      record[idxText],
			Embedding: embedding,
		})

		if len(docs) >= readBatchSize {
			n, err := s.insertInBatches(ctx, docs)
			if err != nil {
				return inserted, scanned, err
			}
			inserted += n
			docs = docs[:0]
		}

		if scanned%10000 == 0 {
			s.log.Info(
				ctx,
				"dataset import progress",
				logger.FieldAny("rows_scanned", scanned),
				logger.FieldAny("rows_inserted", inserted),
				logger.FieldAny("limit", limit),
			)
		}
	}

	if len(docs) > 0 {
		n, err := s.insertInBatches(ctx, docs)
		if err != nil {
			return inserted, scanned, err
		}
		inserted += n
	}

	return inserted, scanned, nil
}

func parseVector(raw string) ([]float32, error) {
	v := strings.TrimSpace(raw)
	v = strings.TrimPrefix(v, "[")
	v = strings.TrimSuffix(v, "]")
	if v == "" {
		return nil, fmt.Errorf("empty vector")
	}

	parts := strings.Split(v, ",")
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		f, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, float32(f))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty vector after parse")
	}
	return out, nil
}

func (s *ImportService) insertInBatches(ctx context.Context, docs []document.Document) (int64, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	writeBatchSize := s.cfg.WriteBatchSize
	if writeBatchSize <= 0 {
		writeBatchSize = 500
	}

	var inserted int64
	for start := 0; start < len(docs); start += writeBatchSize {
		if err := ctx.Err(); err != nil {
			return inserted, err
		}

		end := start + writeBatchSize
		if end > len(docs) {
			end = len(docs)
		}

		n, err := s.docRepo.CreateBatch(ctx, docs[start:end])
		if err != nil {
			return inserted, fmt.Errorf("import service: insert batch: %w", err)
		}
		inserted += n
	}

	return inserted, nil
}

func (s *ImportService) prepareDatasetFile(ctx context.Context) (string, error) {
	if s.cfg.DatasetURL == "" {
		return "", fmt.Errorf("import service: dataset url is empty")
	}
	if s.cfg.LocalPath == "" {
		return "", fmt.Errorf("import service: local path is empty")
	}

	if info, err := os.Stat(s.cfg.LocalPath); err == nil && info.Size() > 0 {
		s.log.Info(ctx, "dataset file already exists, skipping download", logger.FieldAny("path", s.cfg.LocalPath))
		return s.cfg.LocalPath, nil
	}

	if err := os.MkdirAll(filepath.Dir(s.cfg.LocalPath), 0o755); err != nil {
		return "", fmt.Errorf("import service: create dataset dir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.DatasetURL, nil)
	if err != nil {
		return "", fmt.Errorf("import service: create request: %w", err)
	}

	s.log.Info(ctx, "downloading dataset", logger.FieldAny("url", s.cfg.DatasetURL))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("import service: download dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("import service: bad response status: %d", resp.StatusCode)
	}

	file, err := os.Create(s.cfg.LocalPath)
	if err != nil {
		return "", fmt.Errorf("import service: create local dataset file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("import service: save dataset file: %w", err)
	}

	s.log.Info(ctx, "dataset downloaded", logger.FieldAny("path", s.cfg.LocalPath), logger.FieldAny("bytes", written))
	return s.cfg.LocalPath, nil
}

func (s *ImportService) waitForDocumentsTable(ctx context.Context) error {
	const checkEvery = 2 * time.Second

	ticker := time.NewTicker(checkEvery)
	defer ticker.Stop()

	waitLogged := false
	for {
		_, err := s.docRepo.Count(ctx)
		if err == nil {
			if waitLogged {
				s.log.Info(ctx, "import worker: documents table detected, continuing")
			}
			return nil
		}

		if !isUndefinedTableError(err) {
			return fmt.Errorf("import service: check documents table: %w", err)
		}

		if !waitLogged {
			waitLogged = true
			s.log.Info(ctx, "import worker: waiting for migrations, documents table not found")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func isUndefinedTableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
		return true
	}
	return strings.Contains(err.Error(), `relation "documents" does not exist`)
}
