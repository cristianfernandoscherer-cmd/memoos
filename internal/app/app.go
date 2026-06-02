package app

import (
	"context"
	"fmt"

	"github.com/cristian-scherer/memoos/internal/config"
	"github.com/cristian-scherer/memoos/internal/embedding"
	"github.com/cristian-scherer/memoos/internal/logger"
	"github.com/cristian-scherer/memoos/internal/memory"
	"github.com/cristian-scherer/memoos/internal/models"
	"github.com/cristian-scherer/memoos/internal/storage"
)

type App struct {
	Config     *config.Config
	Logger     *logger.Logger
	Storage    *storage.SQLiteStorage
	Embedder   embedding.Embedder
	MemService *memory.Service
	cancel     context.CancelFunc
}

func New() (*App, error) {
	cfg, err := config.Load("./configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log, err := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	st, err := storage.NewSQLiteStorage(cfg.Database.GetDatabasePath(), log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	embedder, err := embedding.NewONNXEmbedder(embedding.ONNXConfig{
		ModelPath: cfg.Embed.ONNXPath,
		ModelName: cfg.Embed.ONNXModel,
		MaxSeqLen: cfg.Embed.MaxSeqLen,
	})
	if err != nil {
		st.Close()
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	memService := memory.NewService(st, embedder, log)

	return &App{
		Config:     cfg,
		Logger:     log,
		Storage:    st,
		Embedder:   embedder,
		MemService: memService,
	}, nil
}

func (a *App) Start() error {
	a.Logger.Info("Starting MemoOS application")
	return nil
}

func (a *App) Stop() error {
	a.Logger.Info("Stopping MemoOS application")

	if a.cancel != nil {
		a.cancel()
	}

	return a.Close()
}

func (a *App) Health(ctx context.Context) error {
	if err := a.Storage.Ping(ctx); err != nil {
		return fmt.Errorf("storage unhealthy: %w", err)
	}

	if err := a.Embedder.IsHealthy(ctx); err != nil {
		return fmt.Errorf("embedder unhealthy: %w", err)
	}

	return nil
}

func (a *App) ClearMemories(ctx context.Context, filter models.MemoryFilter) (int64, error) {
	return a.MemService.ClearMemories(ctx, filter)
}

func (a *App) Close() error {
	var errs []error

	if a.Embedder != nil {
		if err := a.Embedder.Close(); err != nil {
			errs = append(errs, fmt.Errorf("embedder close: %w", err))
		}
	}

	if a.Storage != nil {
		if err := a.Storage.Close(); err != nil {
			errs = append(errs, fmt.Errorf("storage close: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

func (a *App) Context() context.Context {
	if a.cancel == nil {
		a.cancel = func() {}
		return context.Background()
	}
	ctx := context.Background()
	ctx, a.cancel = context.WithCancel(ctx)
	return ctx
}
