package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mini-rag-go/models"
	"net/http"
	"strings"
	"time"
)

// Client Ollama客户端
type Client struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

// NewClient 创建Ollama客户端
func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		Timeout: 60 * time.Second,
	}
}

// Generate 生成文本
func (c *Client) Generate(prompt string, options models.RequestOptions) (string, error) {
	request := models.OllamaRequest{
		Model:   c.Model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败：%v", err)
	}
	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("API请求失败：%v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败：%v", err)
	}
	var response models.OllamaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败：%v", err)
	}
	return response.Response, nil
}

// GenerateStream 流式生成
func (c *Client) GenerateStream(prompt string, options models.RequestOptions, callback func(string)) error {
	request := models.OllamaRequest{
		Model:   c.Model,
		Prompt:  prompt,
		Stream:  true,
		Options: options,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API返回错误：%s - %s", resp.Status, string(body))
	}
	decoder := json.NewDecoder(resp.Body)
	for {
		var response models.OllamaResponse
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("解析流式响应失败：%v", err)
		}
		if response.Response != "" {
			callback(response.Response)
		}
		if response.Done {
			break
		}
	}
	return nil
}

// BuildRAGPrompt 构建RAG提示词
func BuildRAGPrompt(query string, context []models.Document) string {
	var contextBuilder strings.Builder
	//系统指令
	contextBuilder.WriteString("你是一个专业的文档问答助手。请根据提供的文档内容准确回答问题。\n")
	contextBuilder.WriteString("如果文档中没有相关信息，请诚实地告知用户。\n\n")
	// 添加检索到的上下文
	contextBuilder.WriteString("相关文档内容：\n")
	for i, doc := range context {
		contextBuilder.WriteString(fmt.Sprintf("【来源%d:%s】\n", i+1, doc.Filename))
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}
	//用户问题
	contextBuilder.WriteString("基于以上文档内容，请回答以下问题：\n")
	contextBuilder.WriteString(fmt.Sprintf("问题：%s\n\n", query))
	contextBuilder.WriteString("回答：")
	return contextBuilder.String()
}

// BuildRefundPrompt 构建退款相关提示词
func BuildRefundPrompt(query string, context []models.Document) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("你是一个专业的电商客服助手，专门处理退款相关咨询。\n")
	contextBuilder.WriteString("请根据提供的文档信息，清晰、准确地回答用户的退款流程问题。\n\n")
	contextBuilder.WriteString("相关文档信息：\n")

	for i, doc := range context {
		contextBuilder.WriteString(fmt.Sprintf("===== 文档 %d ======\n", i+1))
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}
	contextBuilder.WriteString("用户问题：")
	contextBuilder.WriteString(query)
	contextBuilder.WriteString("\n\n")
	contextBuilder.WriteString("请按照以下要求回答：\n")
	contextBuilder.WriteString("1. 如果文档中有明确的退款流程，请分步骤说明\n")
	contextBuilder.WriteString("2. 如果文档中有时间要求，请明确指出\n")
	contextBuilder.WriteString("3. 如果文档中有联系方式，请提供\n")
	contextBuilder.WriteString("4. 使用友好、专业的语气\n")
	contextBuilder.WriteString("5. 如果文档中没有相关信息，请诚实地告知\n\n")

	contextBuilder.WriteString("回答：")

	return contextBuilder.String()
}

// CheckHealth 检查Ollama服务器是否健康
func (c *Client) CheckHealth() error {
	url := fmt.Sprintf("%s/api/tags", c.BaseURL)

	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("无法连接到Ollama服务：%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama服务返回错误状态码：%d", resp.StatusCode)
	}
	return nil
}
