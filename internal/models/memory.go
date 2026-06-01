package models

import (
	"time"
)

type Memory struct {
	ID int64 `json:"id" db:"id"`

	Project  string  `json:"project" db:"project"`
	Category *string `json:"category" db:"category"`

	Content   string            `json:"content" db:"content"`
	Embedding []float32         `json:"-" db:"embedding"`
	Metadata  map[string]string `json:"metadata" db:"metadata"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MemoryInput struct {
	Project  string            `json:"-"`
	Content  string            `json:"content"`
	Category *string           `json:"category"`
	Metadata map[string]string `json:"metadata"`
	CWD      string            `json:"-"`
}

type SearchResult struct {
	Memory   Memory  `json:"memory"`
	Score    float32 `json:"score"`
	Distance float32 `json:"distance"`
}

type SearchInput struct {
	Project     string
	Query       string
	CWD         string
	Category    *string
	Limit       int
	MinScore    float32
	MaxDistance float32
}

type MemoryFilter struct {
	Project  string
	Category *string
	Since    time.Time
}

type Pagination struct {
	Page     int
	PageSize int
}

func (p *Pagination) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) Limit() int {
	if p.PageSize <= 0 {
		return 10
	}
	return p.PageSize
}

func DefaultPagination() *Pagination {
	return &Pagination{
		Page:     1,
		PageSize: 10,
	}
}
