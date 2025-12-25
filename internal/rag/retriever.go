package rag

import (
	"fmt"
	"mini-rag-go/internal/models"
	"mini-rag-go/internal/store"
	"mini-rag-go/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

// Retriever 检索器
type Retriever struct {
	vectorStore  *store.VectorStore
	chunkSize    int
	chunkOverlap int
}

// NewRetriever 创建检索器
func NewRetriever(store *store.VectorStore, chunkSize, chunkOverlap int) *Retriever {
	return &Retriever{
		vectorStore:  store,
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

// LoadDocumentsFromDir 从目录加载文档
func (r *Retriever) LoadDocumentsFromDir(dirPath string) ([]models.Document, error) {
	var documents []models.Document
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败： %v", err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("警告：无法读取文件安：%s:%v\n", file.Name(), err)
			continue
		}
		doc := models.Document{
			ID:       fmt.Sprintf("%s_%d", file.Name(), len(documents)),
			Content:  string(content),
			Filename: file.Name(),
			Metadata: map[string]string{
				"filename": file.Name(),
				"path":     filePath,
				"type":     "text",
			},
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

// ChunkDocument 分割文档
func (r *Retriever) ChunkDocument(doc models.Document) []models.DocumentChunk {
	var chunks []models.DocumentChunk
	content := doc.Content
	runes := []rune(content)
	totalRunes := len(runes)

	if totalRunes <= r.chunkSize {
		//文档足够小，不需要分割
		chunk := models.DocumentChunk{
			Document: models.Document{
				ID:       fmt.Sprintf("%s_chunk_0", doc.ID),
				Content:  content,
				Filename: doc.Filename,
				Metadata: doc.Metadata,
			},
			ChunkIndex: 0,
			StartPos:   0,
			EndPos:     totalRunes,
		}
		return []models.DocumentChunk{chunk}
	}
	//按句子分割
	sentences := utils.SplitTextBySentences(content)
	var currentChunk strings.Builder
	chunkIndex := 0
	startPos := 0
	for i, sentence := range sentences {
		sentenceRunes := []rune(sentence)

		if currentChunk.Len()+len(sentenceRunes) > r.chunkSize && currentChunk.Len() > 0 {
			//保存当前块
			chunk := models.DocumentChunk{
				Document: models.Document{
					ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
					Content:  currentChunk.String(),
					Filename: doc.Filename,
					Metadata: doc.Metadata,
				},
				ChunkIndex: chunkIndex,
				StartPos:   startPos,
				EndPos:     startPos + len([]rune(currentChunk.String())),
			}
			chunks = append(chunks, chunk)
			//开始新块，保留重叠部分
			chunkIndex++
			currentChunk.Reset()

			//添加重叠的句子（从当前块末尾往回找）
			overlapStart := i - 1
			if overlapStart < 0 {
				overlapStart = 0
			}
			for j := overlapStart; j < i; j++ {
				overlapSentence := sentences[j]
				if len([]rune(overlapSentence)) < r.chunkOverlap {
					currentChunk.WriteString(overlapSentence)
					currentChunk.WriteRune(' ')
				}
			}
			startPos = chunk.EndPos - len([]rune(currentChunk.String()))
		}
		currentChunk.WriteString(sentence)
		currentChunk.WriteString(" ")
	}

	//添加最后一个块
	if currentChunk.Len() > 0 {
		chunk := models.DocumentChunk{
			Document: models.Document{
				ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
				Content:  currentChunk.String(),
				Filename: doc.Filename,
				Metadata: doc.Metadata,
			},
			ChunkIndex: chunkIndex,
			StartPos:   startPos,
			EndPos:     startPos + len([]rune(currentChunk.String())),
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

// Retrieve 检索相关文档
func (r *Retriever) Retrieve(query string, topK int) ([]models.SearchResult, error) {
	return r.vectorStore.Search(query, topK)
}

// BuildVectorStore 构建向量存储
func (r *Retriever) BuildVectorStore(docsPath, storePath string) error {
	//检索是否已存在向量存储
	if _, err := os.Stat(storePath); err == nil {
		fmt.Println("向量存储已存在，跳过构建...")
		return nil
	}
	fmt.Println("正在构建向量存储...")
	//加载文档
	documents, err := r.LoadDocumentsFromDir(docsPath)
	if err != nil {
		return fmt.Errorf("加载文档失败：%v", err)
	}
	fmt.Printf("找到 %d 个文档\n", len(documents))

	//分割文档并添加到向量存储
	totalChunks := 0
	for _, doc := range documents {
		chunks := r.ChunkDocument(doc)
		for _, chunk := range chunks {
			chunkDoc := models.Document{
				ID:       chunk.ID,
				Content:  chunk.Content,
				Filename: chunk.Filename,
				Metadata: chunk.Metadata,
			}
			if err := r.vectorStore.AddDocument(chunkDoc); err != nil {
				fmt.Printf("警告：添加文档块失败：%s：%v", chunk.ID, err)
				continue
			}
			totalChunks++
		}
	}
	fmt.Printf("生成 %d 个文档块\n", totalChunks)

	//保存向量存储
	if err := r.vectorStore.Save(storePath); err != nil {
		return fmt.Errorf("保存向量存储失败：%v", err)
	}
	fmt.Printf("向量存储已保存到 %s\n", storePath)
	return nil
}
