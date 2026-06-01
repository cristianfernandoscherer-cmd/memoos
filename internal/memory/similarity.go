package memory

import (
	"math"
)

// CosineSimilarity computes the cosine similarity between two vectors
// Returns value in [0, 1] where 1 is identical
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// EuclideanDistance computes the Euclidean distance between two vectors
// Returns smaller values for more similar vectors
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}
