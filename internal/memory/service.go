package memory

import (
	"context"
	"sort"
	"time"

	"github.com/cristian-scherer/memoos/internal/embedding"
	"github.com/cristian-scherer/memoos/internal/errors"
	"github.com/cristian-scherer/memoos/internal/logger"
	"github.com/cristian-scherer/memoos/internal/models"
	"github.com/cristian-scherer/memoos/internal/storage"
	"github.com/cristian-scherer/memoos/internal/util"
)

type Service struct {
	storage  storage.Storage
	embedder embedding.Embedder
	logger   *logger.Logger
}

func NewService(st storage.Storage, embedder embedding.Embedder, log *logger.Logger) *Service {
	return &Service{
		storage:  st,
		embedder: embedder,
		logger:   log,
	}
}

func (s *Service) Save(ctx context.Context, input models.MemoryInput) (*models.Memory, error) {
	op := logger.NewOperation(s.logger, "save",
		logger.F("project", input.Project),
		logger.F("category", input.Category),
		logger.F("content_len", len(input.Content)),
	)
	defer op.Success()

	projName := input.Project
	if projName == "" {
		projName = util.ResolvePath(input.CWD)
	}

	if input.Content == "" {
		op.Error(errors.InvalidInput("content is required", "content"))
		return nil, errors.InvalidInput("content is required", "content")
	}
	if len(input.Content) > 10000 {
		op.Error(errors.InvalidInput("content too long (max 10000 chars)", "content"))
		return nil, errors.InvalidInput("content too long (max 10000 chars)", "content")
	}

	embedding, err := s.embedder.Embed(ctx, input.Content)
	if err != nil {
		op.Error(err)
		return nil, errors.EmbeddingError(err, s.embedder.Name())
	}

	mem := models.Memory{
		Project:   projName,
		Category:  input.Category,
		Content:   input.Content,
		Embedding: embedding,
		Metadata:  input.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := s.storage.SaveMemory(ctx, mem)
	if err != nil {
		op.Error(err)
		return nil, errors.DatabaseError(err, "save_memory")
	}

	mem.ID = id
	return &mem, nil
}

func (s *Service) Search(ctx context.Context, input models.SearchInput) ([]models.SearchResult, error) {
	op := logger.NewOperation(s.logger, "search",
		logger.F("project", input.Project),
		logger.F("category", input.Category),
		logger.F("query_len", len(input.Query)),
	)
	defer op.Success()

	projName := input.Project
	if projName == "" {
		projName = util.ResolvePath(input.CWD)
	}

	if input.Query == "" {
		op.Error(errors.InvalidInput("query is required", "query"))
		return nil, errors.InvalidInput("query is required", "query")
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	hasScoreFilter := input.MinScore > 0
	hasDistanceFilter := input.MaxDistance > 0

	if !hasScoreFilter && !hasDistanceFilter {
		input.MinScore = 0.5
	}

	queryEmbedding, err := s.embedder.Embed(ctx, input.Query)
	if err != nil {
		op.Error(err)
		return nil, errors.EmbeddingError(err, s.embedder.Name())
	}

	filter := models.MemoryFilter{
		Project:  projName,
		Category: input.Category,
	}

	results, err := s.storage.SearchMemories(ctx, queryEmbedding, filter, limit*2)
	if err != nil {
		op.Error(err)
		return nil, errors.DatabaseError(err, "search_memories")
	}

	var filtered []models.SearchResult
	for _, r := range results {
		passScoreFilter := !hasScoreFilter || r.Score >= input.MinScore
		passDistanceFilter := !hasDistanceFilter || r.Distance <= input.MaxDistance

		if passScoreFilter && passDistanceFilter {
			filtered = append(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*models.Memory, error) {
	mem, err := s.storage.GetMemory(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to get memory: %v", err)
		return nil, errors.DatabaseError(err, "get_memory")
	}
	if mem == nil {
		return nil, errors.NotFound("memory", id)
	}
	return mem, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	op := logger.NewOperation(s.logger, "delete", logger.F("id", id))
	defer op.Success()

	if err := s.storage.DeleteMemory(ctx, id); err != nil {
		op.Error(err)
		return errors.DatabaseError(err, "delete_memory")
	}
	return nil
}

func (s *Service) List(ctx context.Context, filter models.MemoryFilter, pagination *models.Pagination) ([]models.Memory, error) {
	if pagination == nil {
		pagination = models.DefaultPagination()
	}

	memories, err := s.storage.ListMemories(ctx, filter, pagination)
	if err != nil {
		s.logger.Errorf("failed to list memories: %v", err)
		return nil, errors.DatabaseError(err, "list_memories")
	}

	return memories, nil
}

func (s *Service) ListCategories(ctx context.Context, project string) ([]string, error) {
	categories, err := s.storage.ListCategories(ctx, project)
	if err != nil {
		s.logger.Errorf("failed to list categories: %v", err)
		return nil, errors.DatabaseError(err, "list_categories")
	}
	return categories, nil
}

func (s *Service) Health(ctx context.Context) error {
	if err := s.storage.Ping(ctx); err != nil {
		return errors.UnhealthyError("storage", err)
	}

	if err := s.embedder.IsHealthy(ctx); err != nil {
		return errors.UnhealthyError("embedder", err)
	}

	return nil
}
