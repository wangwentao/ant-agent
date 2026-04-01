package impl

import (
	"context"
	"fmt"
	"os"
)

// ReadFileTool 读取文件的工具
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Reads the content of a file"
}

func (t *ReadFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The path to the file to read",
		},
	}
}

func (t *ReadFileTool) Required() []string {
	return []string{"file_path"}
}

func (t *ReadFileTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	// 检查是否提供了file_path参数
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "Error: file_path parameter is required", nil
	}

	// 检查文件路径是否在当前目录内
	if !isPathInCurrentDir(filePath) {
		return "Error: File operations are restricted to the current directory and its subdirectories", nil
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	return string(content), nil
}
