package impl

import (
	"os"
	"path/filepath"
	"strings"
)

// isPathInCurrentDir 检查文件路径是否在当前目录内
func isPathInCurrentDir(filePath string) bool {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// 清理路径
	cleanPath := filepath.Clean(filePath)

	// 构建绝对路径
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	// 检查路径是否在当前目录内
	return strings.HasPrefix(absPath, cwd)
}
