package vector

import (
	"hash/fnv"
	"mini-rag-go/utils"
	"strings"
)

// Embedder 嵌入器接口
type Embedder interface {
	Embed(text string) ([]float32, error)
	Dimension() int
}

// SimpleEmbedder 简单的嵌入器（基于TF-IDF和n-gram）
type SimpleEmbedder struct {
	dimension int
}

// NewSimpleEmbedder 创建简单嵌入器
func NewSimpleEmbedder(dimension int) *SimpleEmbedder {
	return &SimpleEmbedder{
		dimension: dimension,
	}
}

// Embed 生成嵌入向量
func (e *SimpleEmbedder) Embed(text string) ([]float32, error) {
	text = strings.ToLower(text)
	text = utils.ClearText(text)
	//创建向量
	vector := make([]float32, e.dimension)
	//字符量n-gram特征
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		//1-gram,2-gram,3-gram
		for n := 1; n <= 3 && i+n <= len(runes); n++ {
			ngram := string(runes[i : i+n])
			hash := hashString(ngram) % uint32(e.dimension)
			vector[hash] += 1.0
		}
	}
	//归一化
	utils.NormalizeVector(vector)
	return vector, nil
}

// hashString 字符串哈希
func hashString(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return h.Sum32()
}

// Dimension 返回向量维度
func (e *SimpleEmbedder) Dimension() int {
	return e.dimension
}

// BatchEmbed 批量生成嵌入
func BatchEmbed(embedder Embedder, texts []string) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		vector, err := embedder.Embed(text)
		if err != nil {
			return nil, err
		}
		vectors[i] = vector
	}
	return vectors, nil
}
