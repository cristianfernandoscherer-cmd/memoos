package storage

import (
	"context"
	"time"

	"github.com/cristian-scherer/memoos/internal/models"
)

type Storage interface {
	SaveMemory(ctx context.Context, mem models.Memory) (int64, error)
	GetMemory(ctx context.Context, id int64) (*models.Memory, error)
	ListMemories(ctx context.Context, filter models.MemoryFilter, pagination *models.Pagination) ([]models.Memory, error)
	DeleteMemory(ctx context.Context, id int64) error
	SearchMemories(ctx context.Context, queryEmbedding []float32, filter models.MemoryFilter, limit int) ([]models.SearchResult, error)

	ListCategories(ctx context.Context, project string) ([]string, error)

	Ping(ctx context.Context) error

	Close() error
}

type HealthStatus struct {
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
