package store

import (
	"encoding/json"
	"fmt"
	"mini-rag-go/models"
	"mini-rag-go/utils"
	"mini-rag-go/vector"
	"os"
	"sort"
	"sync"
)

// VectorStore 向量存储
type VectorStore struct {
	documents []models.Document
	vectors   [][]float32
	embedder  vector.Embedder
	mu        sync.RWMutex
}

// NewVectorStore 创建向量存储
func NewVectorStore(embedder vector.Embedder) *VectorStore {
	return &VectorStore{
		documents: make([]models.Document, 0),
		vectors:   make([][]float32, 0),
		embedder:  embedder,
	}
}

// AddDocument 添加文档
func (vs *VectorStore) AddDocument(doc models.Document) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	//生成嵌入
	vector, err := vs.embedder.Embed(doc.Content)
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %v", err)
	}
	vs.documents = append(vs.documents, doc)
	vs.vectors = append(vs.vectors, vector)
	return nil
}

// AddDocuments 批量添加文档
func (vs *VectorStore) AddDocuments(docs []models.Document) error {
	for _, doc := range docs {
		if err := vs.AddDocument(doc); err != nil {
			return err
		}
	}
	return nil
}

// Search 搜索相似文档
func (vs *VectorStore) Search(query string, topK int) ([]models.SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if len(vs.documents) == 0 {
		return []models.SearchResult{}, nil
	}
	//生成查询向量
	queryVector, err := vs.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败：%v", err)
	}
	//计算相似度
	results := make([]models.SearchResult, len(vs.documents))
	for i, vector := range vs.vectors {
		score := utils.CosineSimilarity(queryVector, vector)
		results[i] = models.SearchResult{
			Document: vs.documents[i],
			Score:    score,
		}
	}
	//排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	//返回 topK个结果
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK], nil
}

// Save 保存到文件
func (vs *VectorStore) Save(filename string) error {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	data := struct {
		Documents []models.Document `json:"documents"`
		Vectors   [][]float32       `json:"vectors"`
	}{
		Documents: vs.documents,
		Vectors:   vs.vectors,
	}
	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Errorf("序列化失败：%v", err)
	}
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败：%v", err)
	}
	return nil
}

// Load 从文件加载
func (vs *VectorStore) Load(filename string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件失败：%v", err)
	}
	var storeData struct {
		Documents []models.Document `json:"documents"`
		Vectors   [][]float32       `json:"vectors"`
	}
	if err := json.Unmarshal(data, &storeData); err != nil {
		return fmt.Errorf("反序列化失败：%v", err)
	}
	vs.documents = storeData.Documents
	vs.vectors = storeData.Vectors
	return nil
}

// DocumentCount 返回文档数量
func (vs *VectorStore) DocumentCount() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return len(vs.documents)
}
