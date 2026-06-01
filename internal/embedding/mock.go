package embedding

import (
	"context"
	"math/rand/v2"
)

type MockEmbedder struct {
	*BaseEmbedder
}

func NewMockEmbedder(dimension int) *MockEmbedder {
	return &MockEmbedder{
		BaseEmbedder: NewBaseEmbedder("mock", dimension),
	}
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embedding := make([]float32, m.dimension)
	for i := range embedding {
		embedding[i] = rand.Float32()
	}
	return embedding, nil
}

func (m *MockEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embedding, err := m.Embed(ctx, texts[i])
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}
