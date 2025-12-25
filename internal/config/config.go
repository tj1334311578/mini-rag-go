package config

import (
	"os"
	"strconv"
)

// AppConfig 应用配置
type AppConfig struct {
	DocsPath            string
	VectorStorePath     string
	ChunkSize           int
	ChunkOverlap        int
	TopK                int
	SimilarityThreshold float64
}

// LLMConfig LLM配置
type LLMConfig struct {
	Mode        string
	Model       string
	BaseURL     string
	Temperature float32
	MaxTokens   int
}

// Config 全局配置
type Config struct {
	App AppConfig
	LLM LLMConfig
}

// Global 全局配置实例
var Global *Config

// InitConfig 初始化配置
func InitConfig() {
	//从环境变量读取 LLM模式
	Global = &Config{
		App: AppConfig{
			DocsPath:            getEnv("DOCS_PATH", "docs"),
			VectorStorePath:     getEnv("VECTOR_STORE_PATH", "internal/store/vector_store.json"),
			ChunkSize:           getEnvAsInt("CHUNK_SIZE", 500),
			ChunkOverlap:        getEnvAsInt("CHUNK_OVERLAP", 50),
			TopK:                getEnvAsInt("TOP_K", 3),
			SimilarityThreshold: getEnvAsFloat("SIMILARITY_THRESHOLD", 0.7),
		},
		LLM: LLMConfig{
			Mode:        getEnv("LLM_MODE", os.Getenv("local")),
			Model:       getEnv("OLLAMA_MODEL", os.Getenv("qwen2.5:7b")),
			BaseURL:     getEnv("OLLAMA_BASE_URL", os.Getenv("http://localhost:14434")),
			Temperature: getEnvAsFloat32("LLM_TEMPERATURE", 0.7),
			MaxTokens:   getEnvAsInt("MAX_TOKENS", 1024),
		},
	}
	//打印配置信息
	printConfig()
}

func getEnvAsFloat32(key string, defaultVlue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultVlue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 辅助函数：获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func printConfig() {
	println("=== 配置信息 ===")
	println("文档目录:", Global.App.DocsPath)
	println("向量存储:", Global.App.VectorStorePath)
	println("LLM模式:", Global.LLM.Mode)
	println("LLM模型:", Global.LLM.Model)
	println("Ollama地址:", Global.LLM.BaseURL)
	println("温度:", Global.LLM.Temperature)
	println("===============\n")
}
