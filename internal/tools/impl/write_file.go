package impl

import (
	"context"
	"fmt"
	"os"
)

// WriteFileTool 写入文件的工具
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Writes content to a file"
}

func (t *WriteFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The path to the file to write",
		},
		"content": map[string]interface{}{
			"type":        "string",
			"description": "The content to write to the file",
		},
	}
}

func (t *WriteFileTool) Required() []string {
	return []string{"file_path", "content"}
}

func (t *WriteFileTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	// 检查是否提供了必要参数
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "Error: file_path parameter is required", nil
	}

	// 检查文件路径是否在当前目录内
	if !isPathInCurrentDir(filePath) {
		return "Error: File operations are restricted to the current directory and its subdirectories", nil
	}

	content, ok := input["content"].(string)
	if !ok {
		return "Error: content parameter is required", nil
	}

	// 写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Successfully wrote to file: %s", filePath), nil
}
