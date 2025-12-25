# mini-rag-go

# 运行步骤

# 1. 确保在项目根目录
cd docs-qa-rag

# 2. 安装Ollama（如果还没安装）
curl -fsSL https://ollama.ai/install.sh | sh

# 3. 拉取模型
ollama pull qwen2:0.5b-instruct
# 或者
ollama pull llama3.2:3b

# 4. 启动Ollama服务
ollama serve &
# 或者后台运行
ollama serve > ollama.log 2>&1 &

# 5. 设置环境变量
export LLM_MODE=local
export OLLAMA_MODEL=qwen2:0.5b-instruct

# 6. 运行程序
go run . docs "退款流程是怎样的？"

# 7. 测试其他问题
go run . docs "退款需要多长时间？"
go run . docs "如何联系客服？"
go run . docs "哪些商品不支持退款？"