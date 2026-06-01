package embedding

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	"github.com/yalue/onnxruntime_go"
)

type ONNXConfig struct {
	ModelPath string
	ModelName string
	MaxSeqLen int
}

type ONNXEmbedder struct {
	config        ONNXConfig
	session       *onnxruntime_go.AdvancedSession
	tokenizer     *tokenizer.Tokenizer
	inputIDs      *onnxruntime_go.Tensor[int64]
	attention     *onnxruntime_go.Tensor[int64]
	tokenTypeIDs  *onnxruntime_go.Tensor[int64]
	outputTensor  *onnxruntime_go.Tensor[float32]
	dimension     int
	maxSeqLen     int
	mu            sync.RWMutex
}

func NewONNXEmbedder(
	config ONNXConfig,
) (*ONNXEmbedder, error) {

	if config.MaxSeqLen <= 0 {
		config.MaxSeqLen = 256
	}

	onnxruntime_go.SetSharedLibraryPath(getBundledLibPath())

	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf(
			"failed to initialize ONNX runtime: %w",
			err,
		)
	}

	tokenizerPath := filepath.Join(
		config.ModelPath,
		"tokenizer.json",
	)

	tok, err := pretrained.FromFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load tokenizer from: %s: %w",
			tokenizerPath,
			err,
		)
	}

	inputShape := []int64{
		1,
		int64(config.MaxSeqLen),
	}

	inputIDs, err := onnxruntime_go.NewTensor(
		inputShape,
		make([]int64, config.MaxSeqLen),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create input_ids tensor: %w",
			err,
		)
	}

	attention, err := onnxruntime_go.NewTensor(
		inputShape,
		make([]int64, config.MaxSeqLen),
	)

	if err != nil {
		inputIDs.Destroy()

		return nil, fmt.Errorf(
			"failed to create attention tensor: %w",
			err,
		)
	}

	tokenTypeIDs, err := onnxruntime_go.NewTensor(
		inputShape,
		make([]int64, config.MaxSeqLen),
	)

	if err != nil {
		inputIDs.Destroy()
		attention.Destroy()

		return nil, fmt.Errorf(
			"failed to create token_type_ids tensor: %w",
			err,
		)
	}

	outputShape := []int64{
		1,
		int64(config.MaxSeqLen),
		384,
	}

	outputTensor, err := onnxruntime_go.NewTensor(
		outputShape,
		make([]float32, config.MaxSeqLen*384),
	)

	if err != nil {
		inputIDs.Destroy()
		attention.Destroy()
		tokenTypeIDs.Destroy()

		return nil, fmt.Errorf(
			"failed to create output tensor: %w",
			err,
		)
	}

	modelPath := filepath.Join(
		config.ModelPath,
		"model.onnx",
	)

	session, err := onnxruntime_go.NewAdvancedSession(
		modelPath,

		[]string{
			"input_ids",
			"attention_mask",
			"token_type_ids",
		},

		[]string{
			"last_hidden_state",
		},

		[]onnxruntime_go.Value{
			inputIDs,
			attention,
			tokenTypeIDs,
		},

		[]onnxruntime_go.Value{
			outputTensor,
		},

		nil,
	)

	if err != nil {
		inputIDs.Destroy()
		attention.Destroy()
		tokenTypeIDs.Destroy()
		outputTensor.Destroy()

		return nil, fmt.Errorf(
			"failed to create ONNX session: %w",
			err,
		)
	}

	return &ONNXEmbedder{
		config:       config,
		session:      session,
		tokenizer:    tok,
		inputIDs:     inputIDs,
		attention:    attention,
		tokenTypeIDs: tokenTypeIDs,
		outputTensor: outputTensor,
		dimension:    384,
		maxSeqLen:    config.MaxSeqLen,
	}, nil
}

func (o *ONNXEmbedder) Embed(
	ctx context.Context,
	text string,
) ([]float32, error) {

	o.mu.Lock()
	defer o.mu.Unlock()

	encoding, err := o.tokenizer.EncodeSingle(text, true)
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize: %w", err)
	}

	ids := encoding.Ids
	mask := encoding.AttentionMask

	inputIDs := o.inputIDs.GetData()
	attention := o.attention.GetData()
	tokenTypeIDs := o.tokenTypeIDs.GetData()

	for i := 0; i < o.maxSeqLen; i++ {
		tokenTypeIDs[i] = 0
		if i < len(ids) {
			inputIDs[i] = int64(ids[i])
			attention[i] = int64(mask[i])
		} else {
			inputIDs[i] = 0
			attention[i] = 0
		}
	}

	if err := o.session.Run(); err != nil {
		return nil, fmt.Errorf(
			"failed to run inference: %w",
			err,
		)
	}

	output := o.outputTensor.GetData()

	embedding := o.meanPool(
		output,
		attention,
	)

	o.normalize(embedding)

	return embedding, nil
}

func (o *ONNXEmbedder) BatchEmbed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {

	results := make([][]float32, 0, len(texts))

	for _, text := range texts {

		embedding, err := o.Embed(ctx, text)
		if err != nil {
			return nil, err
		}

		results = append(results, embedding)
	}

	return results, nil
}

func (o *ONNXEmbedder) meanPool(
	output []float32,
	attention []int64,
) []float32 {

	embedding := make([]float32, o.dimension)

	var validTokens float32

	for tokenIdx := 0; tokenIdx < o.maxSeqLen; tokenIdx++ {

		if attention[tokenIdx] == 0 {
			continue
		}

		offset := tokenIdx * o.dimension

		for j := 0; j < o.dimension; j++ {
			embedding[j] += output[offset+j]
		}

		validTokens++
	}

	if validTokens == 0 {
		return embedding
	}

	for j := 0; j < o.dimension; j++ {
		embedding[j] /= validTokens
	}

	return embedding
}

func (o *ONNXEmbedder) normalize(
	v []float32,
) {

	var sum float32

	for _, x := range v {
		sum += x * x
	}

	norm := float32(math.Sqrt(float64(sum)))

	if norm == 0 {
		return
	}

	for i := range v {
		v[i] /= norm
	}
}

func (o *ONNXEmbedder) Name() string {
	return "onnx"
}

func (o *ONNXEmbedder) Dimension() int {
	return o.dimension
}

func (o *ONNXEmbedder) IsHealthy(
	ctx context.Context,
) error {

	if o.session == nil {
		return fmt.Errorf("session is nil")
	}

	return nil
}

func (o *ONNXEmbedder) Close() error {

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.session != nil {
		o.session.Destroy()
	}

	if o.inputIDs != nil {
		o.inputIDs.Destroy()
	}

	if o.attention != nil {
		o.attention.Destroy()
	}

	if o.tokenTypeIDs != nil {
		o.tokenTypeIDs.Destroy()
	}

	if o.outputTensor != nil {
		o.outputTensor.Destroy()
	}

	return nil
}

func getBundledLibPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "libs/libonnxruntime.dylib"
	case "windows":
		return "libs/onnxruntime.dll"
	default:
		return "libs/libonnxruntime.so"
	}
}
