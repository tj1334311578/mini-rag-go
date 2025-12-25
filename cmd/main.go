package main

import (
	"fmt"
	"log"
	"mini-rag-go/internal/config"
	"mini-rag-go/internal/models"
	"mini-rag-go/internal/ollama"
	rag2 "mini-rag-go/internal/rag"
	"mini-rag-go/internal/store"
	"mini-rag-go/internal/vector"
	"os"
	"strings"
)

func main() {
	//ans, _ := llm.Ask("用一句话解释什么是 RAG")
	//fmt.Println(ans)
	//1.初始化配置
	config.InitConfig()
	cfg := config.Global
	fmt.Println("🎯 文档问答RAG系统")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("模式: %s | 模型: %s\n", cfg.LLM.Mode, cfg.LLM.Model)
	fmt.Println(strings.Repeat("=", 50))
	//2.检查参数
	if len(os.Args) < 3 {
		printUsage()
		return
	}
	command := os.Args[1]
	query := strings.Join(os.Args[2:], " ")
	if command != "docs" {
		fmt.Println("❌ 未知命令，请使用 'docs'")
		printUsage()
		return
	}
	// 3.初始化组件
	fmt.Println("🔄 初始化系统组件...")
	//创建嵌入器
	embedder := vector.NewSimpleEmbedder(300)
	//创建向量存储
	vectorStore := store.NewVectorStore(embedder)
	//创建检索器
	retriever := rag2.NewRetriever(vectorStore, cfg.App.ChunkSize, cfg.App.ChunkOverlap)
	//4.检查或构建向量存储
	vectorStorePath := cfg.App.VectorStorePath
	if _, err := os.Stat(vectorStorePath); os.IsNotExist(err) {
		fmt.Println("📚 构建向量存储...")
		if err := retriever.BuildVectorStore(cfg.App.DocsPath, vectorStorePath); err != nil {
			log.Fatalf("❌ 构建向量存储失败: %v", err)
		}
	} else {
		fmt.Println("📖 加载现有向量存储...")
		if err := vectorStore.Load(vectorStorePath); err != nil {
			log.Fatalf("❌ 加载向量存储失败: %v", err)
		}
		fmt.Printf("✅ 已加载 %d 个文档块\n", vectorStore.DocumentCount())
	}
	// 5.处理查询
	fmt.Printf("\n❓ 问题: %s\n", query)
	fmt.Println("🔍 检索相关文档...")
	searchResults, err := retriever.Retrieve(query, cfg.App.TopK)
	if err != nil {
		log.Fatalf("❌ 检索失败: %v", err)
	}
	if len(searchResults) == 0 {
		fmt.Println("❌ 未找到相关文档")
		return
	}
	fmt.Printf("✅ 找到 %d 个相关文档片段\n", len(searchResults))

	//6.生成回答
	var answer string
	if cfg.LLM.Model == "local" {
		//检查 Ollama服务
		fmt.Println("🧠 检查Ollama服务...")
		ollamaClient := ollama.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model)

		if err := ollamaClient.CheckHealth(); err != nil {
			fmt.Printf("⚠️  Ollama服务不可用: %v\n", err)
			fmt.Println("🔄 切换到降级模式...")
			answer = generateFallbackAnswer(query, searchResults)
		} else {
			fmt.Println("✅ Ollama服务正常，生成回答...")
			generator := rag2.NewGenerator(ollamaClient)
			answer, err = generator.GenerateAnswer(query, searchResults)
			if err != nil {
				fmt.Printf("⚠️  LLM生成失败: %v\n", err)
				answer = generateFallbackAnswer(query, searchResults)
			}
		}
	} else {
		//使用降级模式
		fmt.Println("📝 使用规则引擎生成回答...")
		answer = generateFallbackAnswer(query, searchResults)
	}
	//7.显示结果
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("💡 回答:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(answer)
	fmt.Println(strings.Repeat("-", 50))

	//8.显示来源
	if len(searchResults) > 0 {
		fmt.Println("\n📚 参考来源:")
		for i, result := range searchResults {
			content := result.Document.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("%d. [%s] (相似度: %.2f)\n   %s\n", i+1, result.Document.Filename, result.Score, content)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
}

// generateFallbackAnswer 生成降级回答
func generateFallbackAnswer(query string, results []models.SearchResult) string {
	var answer strings.Builder
	answer.WriteString("根据文档内容：\n\n")
	for i, result := range results {
		//简单提取相关信息
		content := extractRelevantInfo(result.Document.Content, query)
		if content != "" {
			answer.WriteString(fmt.Sprintf("%d,%s\n\n", i+1, content))
		}
	}
	if answer.Len() == 0 {
		return "抱歉，没有找到明确的相关信息。"
	}
	return answer.String()
}

// extractRelevantInfo 提取相关信息
func extractRelevantInfo(content, query string) string {
	lines := strings.Split(content, "\n")
	var relevantLines []string
	lowerQuery := strings.ToLower(query)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		//简单的关键词匹配
		if strings.Contains(strings.ToLower(line), "退款") && strings.Contains(lowerQuery, "退款") {
			relevantLines = append(relevantLines, line)
		} else if strings.Contains(strings.ToLower(line), "流程") && (strings.Contains(lowerQuery, "流程") || strings.Contains(lowerQuery, "步骤")) {
			relevantLines = append(relevantLines, line)
		} else if strings.Contains(strings.ToLower(line), "时间") && strings.Contains(lowerQuery, "时间") {
			relevantLines = append(relevantLines, line)
		} else if strings.Contains(strings.ToLower(line), "联系") && strings.Contains(lowerQuery, "联系") {
			relevantLines = append(relevantLines, line)
		}
	}
	if len(relevantLines) > 0 {
		return strings.Join(relevantLines, "\n")
	}
	//如果没有匹配的关键词，返回前两句
	if len(lines) >= 2 {
		return lines[0] + "\n" + lines[1]
	} else if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

// printUsage 打印使用方法
func printUsage() {
	fmt.Println("使用方法:")
	fmt.Println("  go run . docs \"你的问题\"")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  export LLM_MODE=local")
	fmt.Println("  export OLLAMA_MODEL=qwen2:0.5b-instruct")
	fmt.Println("  go run . docs \"退款流程是怎样的？\"")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  LLM_MODE          本地模式: local (默认)")
	fmt.Println("  OLLAMA_MODEL      Ollama模型名称")
	fmt.Println("  OLLAMA_BASE_URL   Ollama服务地址")
	fmt.Println("  DOCS_PATH         文档目录路径")
}
