package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mini-rag-go/models"
	"net/http"
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
