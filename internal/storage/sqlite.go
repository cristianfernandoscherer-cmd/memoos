package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"github.com/cristian-scherer/memoos/internal/logger"
	"github.com/cristian-scherer/memoos/internal/models"
)

type SQLiteStorage struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewSQLiteStorage(dbPath string, dbLogger *logger.Logger) (*SQLiteStorage, error) {
	dsn := dbPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{
		db:     db,
		logger: dbLogger,
	}

	if err := storage.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return storage, nil
}

func NewSQLiteStorageWithDB(db *sql.DB, dbLogger *logger.Logger) *SQLiteStorage {
	return &SQLiteStorage{
		db:     db,
		logger: dbLogger,
	}
}

func (s *SQLiteStorage) Migrate(ctx context.Context) error {
	schema, err := os.ReadFile("internal/storage/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SaveMemory(ctx context.Context, mem models.Memory) (int64, error) {
	embeddingJSON, err := json.Marshal(mem.Embedding)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	metadataJSON, err := json.Marshal(mem.Metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now()
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = now
	}
	mem.UpdatedAt = now

	var result sql.Result
	if mem.Category == nil {
		result, err = s.db.ExecContext(ctx, `
			INSERT INTO memories (project, category, content, embedding, metadata, created_at, updated_at)
			VALUES (?, NULL, ?, ?, ?, ?, ?)
		`, mem.Project, mem.Content, string(embeddingJSON),
			string(metadataJSON), mem.CreatedAt, mem.UpdatedAt)
	} else {
		result, err = s.db.ExecContext(ctx, `
			INSERT INTO memories (project, category, content, embedding, metadata, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, mem.Project, mem.Category, mem.Content, string(embeddingJSON),
			string(metadataJSON), mem.CreatedAt, mem.UpdatedAt)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to insert memory: %w", err)
	}

	return result.LastInsertId()
}

func (s *SQLiteStorage) GetMemory(ctx context.Context, id int64) (*models.Memory, error) {
	var mem models.Memory
	var embeddingJSON, metadataJSON string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, project, category, content, embedding, metadata, created_at, updated_at
		FROM memories WHERE id = ?
	`, id).Scan(
		&mem.ID, &mem.Project, &mem.Category, &mem.Content,
		&embeddingJSON, &metadataJSON,
		&mem.CreatedAt, &mem.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query memory: %w", err)
	}

	if err := json.Unmarshal([]byte(embeddingJSON), &mem.Embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &mem.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &mem, nil
}

func (s *SQLiteStorage) ListMemories(ctx context.Context, filter models.MemoryFilter, pagination *models.Pagination) ([]models.Memory, error) {
	query := `
		SELECT id, project, category, content, embedding, metadata, created_at, updated_at
		FROM memories WHERE 1=1
	`
	args := []interface{}{}

	if filter.Project != "" {
		query += " AND project = ?"
		args = append(args, filter.Project)
	}
	if filter.Category != nil {
		query += " AND category = ?"
		args = append(args, *filter.Category)
	}
	if !filter.Since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Since)
	}

	query += " ORDER BY created_at DESC"
	if pagination != nil {
		query += " LIMIT ? OFFSET ?"
		args = append(args, pagination.Limit(), pagination.Offset())
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var memories []models.Memory
	for rows.Next() {
		var mem models.Memory
		var embeddingJSON, metadataJSON string

		if err := rows.Scan(
			&mem.ID, &mem.Project, &mem.Category, &mem.Content,
			&embeddingJSON, &metadataJSON,
			&mem.CreatedAt, &mem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &mem.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}
		if err := json.Unmarshal([]byte(metadataJSON), &mem.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		memories = append(memories, mem)
	}

	return memories, nil
}

func (s *SQLiteStorage) DeleteMemory(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil
	}

	return nil
}

func (s *SQLiteStorage) SearchMemories(ctx context.Context, queryEmbedding []float32, filter models.MemoryFilter, limit int) ([]models.SearchResult, error) {

	query := `
		SELECT id, project, COALESCE(category, '') as category, content, embedding, metadata, created_at, updated_at
		FROM memories WHERE 1=1
	`
	args := []interface{}{}

	if filter.Project != "" {
		query += " AND project = ?"
		args = append(args, filter.Project)
	}
	if filter.Category != nil {
		query += " AND category = ?"
		args = append(args, *filter.Category)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories for search: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var mem models.Memory
		var embeddingJSON, metadataJSON string
		var category string

		if err := rows.Scan(
			&mem.ID, &mem.Project, &category, &mem.Content,
			&embeddingJSON, &metadataJSON,
			&mem.CreatedAt, &mem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}

		if category != "" {
			mem.Category = &category
		} else {
			mem.Category = nil
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &mem.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}
		if err := json.Unmarshal([]byte(metadataJSON), &mem.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		score := models.CosineSimilarity(queryEmbedding, mem.Embedding)
		distance := models.EuclideanDistance(queryEmbedding, mem.Embedding)

		results = append(results, models.SearchResult{
			Memory:   mem,
			Score:    score,
			Distance: distance,
		})
	}

	return results, nil
}

func (s *SQLiteStorage) ListCategories(ctx context.Context, project string) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM memories
		WHERE project = ?
		ORDER BY category
	`

	rows, err := s.db.QueryContext(ctx, query, project)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *SQLiteStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLiteStorage) ClearMemories(ctx context.Context, filter models.MemoryFilter) (int64, error) {
	query := "DELETE FROM memories WHERE 1=1"
	args := []interface{}{}

	if filter.Project != "" {
		query += " AND project = ?"
		args = append(args, filter.Project)
	}
	if filter.Category != nil {
		query += " AND category = ?"
		args = append(args, *filter.Category)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to clear memories: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
