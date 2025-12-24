package utils

import (
	"math"
	"strings"
	"unicode"
)

// Min 返回最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TruncateText 截断文本
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength]) + "..."
}

// CosineSimilarity 计算余弦相似度
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	var dotProduct float64
	var normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// NormalizeVector 归一化向量
func NormalizeVector(vector []float32) {
	var sum float64
	for _, v := range vector {
		sum += float64(v * v)
	}
	if sum > 0 {
		norm := float32(math.Sqrt(sum))
		for i := range vector {
			vector[i] /= norm
		}
	}
}

// SplitTextBySentences 按句子分割文本
func SplitTextBySentences(text string) []string {
	var sentences []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		//简单的中英文句子结束判断
		if r == '。' || r == '！' || r == '？' ||
			r == '.' || r == '!' || r == '?' {
			// 检查是否是缩写
			if current.Len() > 1 {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}
	return sentences
}

// ClearText 清理文本
func ClearText(text string) string {
	//移除多余空白字符
	text = strings.Join(strings.Fields(text), " ")
	//移除控制字符
	var result strings.Builder
	for _, r := range text {
		if unicode.IsGraphic(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ContainsChinese 检查是否包含中文
func ContainsChinese(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
