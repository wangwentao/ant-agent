package impl

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EditFileTool 编辑文件的工具
type EditFileTool struct{}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Description() string {
	return "Edits a file by replacing a string"
}

func (t *EditFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "The path to the file to edit",
		},
		"old_string": map[string]interface{}{
			"type":        "string",
			"description": "The string to replace in the file",
		},
		"new_string": map[string]interface{}{
			"type":        "string",
			"description": "The new string to replace with",
		},
	}
}

func (t *EditFileTool) Required() []string {
	return []string{"file_path", "old_string", "new_string"}
}

func (t *EditFileTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	// 检查是否提供了必要参数
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "Error: file_path parameter is required", nil
	}

	// 检查文件路径是否在当前目录内
	if !isPathInCurrentDir(filePath) {
		return "Error: File operations are restricted to the current directory and its subdirectories", nil
	}

	oldString, ok := input["old_string"].(string)
	if !ok {
		return "Error: old_string parameter is required", nil
	}

	newString, ok := input["new_string"].(string)
	if !ok {
		return "Error: new_string parameter is required", nil
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	// 替换内容
	originalContent := string(content)
	updatedContent := strings.ReplaceAll(originalContent, oldString, newString)

	// 写入更新后的内容
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Successfully edited file: %s\nReplaced '%s' with '%s'", filePath, oldString, newString), nil
}
