package rag

import (
	"fmt"
	"mini-rag-go/internal/config"
	models2 "mini-rag-go/internal/models"
	"mini-rag-go/internal/ollama"
	"mini-rag-go/internal/utils"
	"strings"
)

// Generator 回答生成器
type Generator struct {
	ollamaClient *ollama.Client
}

// NewGenerator 创建生成器
func NewGenerator(client *ollama.Client) *Generator {
	return &Generator{
		ollamaClient: client,
	}
}

// GenerateAnswer 生成回答
func (g *Generator) GenerateAnswer(query string, searchResults []models2.SearchResult) (string, error) {
	if len(searchResults) == 0 {
		return "抱歉，没有找到相关信息。", nil
	}
	//转换为 Document切片
	documents := make([]models2.Document, len(searchResults))
	for i, result := range searchResults {
		documents[i] = result.Document
	}
	// 根据查询类型选择提示词模板
	var prompt string
	if strings.Contains(query, "退款") || strings.Contains(query, "退货") {
		prompt = ollama.BuildRefundPrompt(query, documents)
	} else {
		prompt = ollama.BuildRAGPrompt(query, documents)
	}
	// 设置生成选项
	options := models2.RequestOptions{
		Temperature: config.Global.LLM.Temperature,
		TopP:        0.9,
		TopK:        40,
		NumPredict:  config.Global.LLM.MaxTokens,
	}
	//调用 Ollama生成回答
	answer, err := g.ollamaClient.Generate(prompt, options)
	if err != nil {
		return "", fmt.Errorf("生成回答失败：%v", err)
	}
	//清理回答
	answer = strings.TrimSpace(answer)
	return answer, err
}

// GenerateAnswerWithFallback 带降级的回答生成
func (g *Generator) GenerateAnswerWithFallback(query string, searchResults []models2.SearchResult) string {
	// 首先尝试使用 LLM生成
	if config.Global.LLM.Mode == "local" {
		answer, err := g.GenerateAnswer(query, searchResults)
		if err == nil && answer != "" {
			return answer
		}
		fmt.Printf("LLM生成失败，使用降级方案：%v\n", err)
	}
	//降级方案：基于规则的生成
	return generateRuleBaseAnswer(query, searchResults)
}

// generateRuleBasedAnswer 基于规则的生成（降级方案）
func generateRuleBaseAnswer(query string, searchResults []models2.SearchResult) string {
	var answer strings.Builder

	if len(searchResults) == 0 {
		return "抱歉，没有找到相关信息。"
	}
	//根据查询类型生成不同格式的回答
	lowerQuery := strings.ToLower(query)
	if strings.Contains(lowerQuery, "流程") || strings.Contains(lowerQuery, "步骤") || strings.Contains(lowerQuery, "怎么") || strings.Contains(lowerQuery, "如何") {
		answer.WriteString("根据文档内容，相关流程如下：\n\n")
		for i, result := range searchResults {
			content := extractProcessSteps(result.Document.Content)
			if content != "" {
				answer.WriteString(fmt.Sprintf("%d. %s\n", i+1,
					utils.TruncateText(content, 200)))
			}
		}
	} else if strings.Contains(lowerQuery, "时间") || strings.Contains(lowerQuery, "多久") {
		answer.WriteString("根据文档中的时间信息：\n\n")
		for _, result := range searchResults {
			content := extractTimeInfo(result.Document.Content)
			if content != "" {
				answer.WriteString(fmt.Sprintf("• %s\n", content))
			}
		}
	} else {
		// 通用回答
		answer.WriteString("根据文档信息：\n\n")

		for i, result := range searchResults {
			answer.WriteString(fmt.Sprintf("%d. %s\n\n", i+1,
				utils.TruncateText(result.Document.Content, 150)))
		}
	}
	if answer.Len() == 0 {
		answer.WriteString("文档中没有找到明确的相关信息。")
	}
	return answer.String()
}

// extractProcessSteps 提取流程步骤
func extractProcessSteps(content string) string {
	var steps []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		//匹配步骤格式
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") ||
			strings.HasPrefix(line, "5.") || strings.HasPrefix(line, "6.") ||
			strings.HasPrefix(line, "a.") || strings.HasPrefix(line, "b.") ||
			strings.HasPrefix(line, "c.") || strings.HasPrefix(line, "d.") ||
			strings.Contains(line, "第一步") || strings.Contains(line, "第二步") ||
			strings.Contains(line, "登录") || strings.Contains(line, "进入") ||
			strings.Contains(line, "选择") || strings.Contains(line, "点击") ||
			strings.Contains(line, "提交") || strings.Contains(line, "等待") {

			steps = append(steps, line)
		}
	}
	if len(steps) > 0 {
		return strings.Join(steps, "\n")
	}
	// 如果没有明确的步骤，返回相关内容
	sentences := utils.SplitTextBySentences(content)
	if len(sentences) > 0 {
		return sentences[0]
	}
	return ""
}

// extractTimeInfo 提取时间信息
func extractTimeInfo(content string) string {
	var timeInfo []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "工作日") || strings.Contains(line, "小时") ||
			strings.Contains(line, "天") || strings.Contains(line, "分钟") ||
			strings.Contains(line, "时间") || strings.Contains(line, "审核") ||
			strings.Contains(line, "到账") || strings.Contains(line, "期限") {

			timeInfo = append(timeInfo, line)
		}
	}

	if len(timeInfo) > 0 {
		return strings.Join(timeInfo, "; ")
	}

	return ""
}

// extractContactInfo 提取联系信息
func extractContactInfo(content string) string {
	var contacts []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "@") || strings.Contains(line, "邮箱") ||
			strings.Contains(line, "电话") || strings.Contains(line, "客服") ||
			strings.Contains(line, "400-") || strings.Contains(line, "微信") ||
			strings.Contains(line, "QQ") {

			contacts = append(contacts, line)
		}
	}

	if len(contacts) > 0 {
		return strings.Join(contacts, "; ")
	}

	return ""
}
