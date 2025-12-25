package models

// Document 文档结构
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Filename  string            `json:"filename"`
	Metadata  map[string]string `json:"metadata"`
	Embedding []float32         `json:"embedding"`
}

// DocumentChunk 文档分块
type DocumentChunk struct {
	Document
	ChunkIndex int `json:"chunk_index"`
	StartPos   int `json:"start_pos"`
	EndPos     int `json:"end_pos"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Document Document
	Score    float64
}
