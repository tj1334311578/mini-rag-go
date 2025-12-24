package models

// OllamaRequest Ollama API 请求
type OllamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options RequestOptions `json:"options"`
}
type RequestOptions struct {
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// OllamaResponse Ollama API响应
type OllamaResponse struct {
	Model      string `json:"model"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"` //system，user，assistant
	Content string `json:"content"`
}
