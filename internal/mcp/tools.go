package mcp

import (
	"context"
	"encoding/json"

	"github.com/cristian-scherer/memoos/internal/memory"
	"github.com/cristian-scherer/memoos/internal/models"
)

type ToolHandler struct {
	memoryService *memory.Service
}

func NewToolHandler(memService *memory.Service) *ToolHandler {
	return &ToolHandler{
		memoryService: memService,
	}
}

type SaveMemoryInput struct {
	CWD      string            `json:"cwd"`
	Category *string           `json:"category"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

type SaveMemoryOutput struct {
	Success bool           `json:"success"`
	Memory  *models.Memory `json:"memory,omitempty"`
	Error   *ErrorDetail   `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type SearchMemoryInput struct {
	CWD         string  `json:"cwd"`
	Query       string  `json:"query"`
	Category    *string `json:"category"`
	Limit       int     `json:"limit"`
	MinScore    float32 `json:"min_score"`
	MaxDistance float32 `json:"max_distance"`
}

type SearchMemoryOutput struct {
	Success                 bool                  `json:"success"`
	Results                 []models.SearchResult `json:"results,omitempty"`
	QueryEmbeddingDimension int                   `json:"query_embedding_dimension,omitempty"`
	TotalMemoriesSearched   int                   `json:"total_memories_searched,omitempty"`
	Error                   *ErrorDetail          `json:"error,omitempty"`
}

type ListMemoryInput struct {
	CWD      string `json:"cwd"`
	Category string `json:"category"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type ListMemoryOutput struct {
	Success  bool            `json:"success"`
	Memories []models.Memory `json:"memories,omitempty"`
	Total    int             `json:"total,omitempty"`
	Error    *ErrorDetail    `json:"error,omitempty"`
}

func (h *ToolHandler) HandleSaveMemory(ctx context.Context, input SaveMemoryInput) SaveMemoryOutput {
	var category *string
	if input.Category != nil {
		category = input.Category
	}

	memInput := models.MemoryInput{
		CWD:      input.CWD,
		Category: category,
		Content:  input.Content,
		Metadata: input.Metadata,
	}

	mem, err := h.memoryService.Save(ctx, memInput)
	if err != nil {
		return SaveMemoryOutput{
			Success: false,
			Error: &ErrorDetail{
				Code:    "ERROR",
				Message: err.Error(),
			},
		}
	}

	return SaveMemoryOutput{
		Success: true,
		Memory:  mem,
	}
}

func (h *ToolHandler) HandleSearchMemory(ctx context.Context, input SearchMemoryInput) SearchMemoryOutput {
	var category *string
	if input.Category != nil {
		category = input.Category
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	minScore := input.MinScore
	if minScore <= 0 {
		minScore = 0.5
	}

	searchInput := models.SearchInput{
		CWD:         input.CWD,
		Query:       input.Query,
		Category:    category,
		Limit:       limit,
		MinScore:    minScore,
		MaxDistance: input.MaxDistance,
	}

	results, err := h.memoryService.Search(ctx, searchInput)
	if err != nil {
		return SearchMemoryOutput{
			Success: false,
			Error: &ErrorDetail{
				Code:    "ERROR",
				Message: err.Error(),
			},
		}
	}

	return SearchMemoryOutput{
		Success:                 true,
		Results:                 results,
		QueryEmbeddingDimension: len(results),
		TotalMemoriesSearched:   len(results),
	}
}

func (h *ToolHandler) HandleListMemory(ctx context.Context, input ListMemoryInput) ListMemoryOutput {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	var category *string
	if input.Category != "" {
		category = &input.Category
	}

	filter := models.MemoryFilter{
		Category: category,
		Project:  "",
	}

	pagination := &models.Pagination{
		Page:     (offset / limit) + 1,
		PageSize: limit,
	}

	memories, err := h.memoryService.List(ctx, filter, pagination)
	if err != nil {
		return ListMemoryOutput{
			Success: false,
			Error: &ErrorDetail{
				Code:    "ERROR",
				Message: err.Error(),
			},
		}
	}

	return ListMemoryOutput{
		Success:  true,
		Memories: memories,
		Total:    len(memories),
	}
}

func (h *ToolHandler) RegisterTools(server *Server) {
	server.RegisterTool(Tool{
		Name:        "memory_save",
		Description: "Save a memory to the semantic store",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "Current working directory for project resolution",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Memory category for context separation (optional, NULL if empty)",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Memory content to save",
					"minLength":   1,
					"maxLength":   10000,
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Optional metadata key-value pairs",
				},
			},
			"required": []string{"cwd", "content"},
		},
		Handler: func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var in SaveMemoryInput
			if err := json.Unmarshal(input, &in); err != nil {
				return nil, err
			}
			return h.HandleSaveMemory(ctx, in), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "memory_search",
		Description: "Search memories by semantic similarity",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "Current working directory for project resolution",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
					"minLength":   1,
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Filter by category (optional)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default: 10, max: 100)",
					"minimum":     1,
					"maximum":     100,
				},
				"min_score": map[string]interface{}{
					"type":        "number",
					"description": "Minimum similarity score (0.0-1.0)",
					"minimum":     0.0,
					"maximum":     1.0,
				},
				"max_distance": map[string]interface{}{
					"type":        "number",
					"description": "Maximum Euclidean distance",
					"minimum":     0.0,
				},
			},
			"required": []string{"cwd", "query"},
		},
		Handler: func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var in SearchMemoryInput
			if err := json.Unmarshal(input, &in); err != nil {
				return nil, err
			}
			return h.HandleSearchMemory(ctx, in), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "memory_list",
		Description: "List recent memories",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "Current working directory for project resolution",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Filter by category (optional)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default: 20)",
					"minimum":     1,
					"maximum":     100,
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Pagination offset",
					"minimum":     0,
				},
			},
			"required": []string{"cwd"},
		},
		Handler: func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var in ListMemoryInput
			if err := json.Unmarshal(input, &in); err != nil {
				return nil, err
			}
			return h.HandleListMemory(ctx, in), nil
		},
	})
}
