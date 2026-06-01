package embedding

import "context"

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
	Name() string
	Dimension() int
	Close() error
	IsHealthy(ctx context.Context) error
}

type BaseEmbedder struct {
	name      string
	dimension int
}

func NewBaseEmbedder(name string, dimension int) *BaseEmbedder {
	return &BaseEmbedder{
		name:      name,
		dimension: dimension,
	}
}

func (b *BaseEmbedder) Name() string {
	return b.name
}

func (b *BaseEmbedder) Dimension() int {
	return b.dimension
}

func (b *BaseEmbedder) IsHealthy(ctx context.Context) error {
	return nil
}

func (b *BaseEmbedder) Close() error {
	return nil
}
