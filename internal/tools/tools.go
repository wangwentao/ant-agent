package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ant-agent/internal/skills"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input map[string]interface{}) (string, error)
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建一个新的工具注册表
func NewToolRegistry(skillCatalog *skills.SkillCatalog) *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]Tool),
	}

	// 注册内置工具
	registry.Register(&ShellCommandTool{})
	registry.Register(&ReadFileTool{})
	registry.Register(&WriteFileTool{})
	registry.Register(&EditFileTool{})
	registry.Register(&ListSkillsTool{skillCatalog: skillCatalog})
	registry.Register(&ActivateSkillTool{skillCatalog: skillCatalog})

	return registry
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// GetTool 根据名称获取工具
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools 获取所有工具
func (r *ToolRegistry) GetAllTools() map[string]Tool {
	return r.tools
}

// RunTool 运行工具
func (r *ToolRegistry) RunTool(ctx context.Context, toolName string, input map[string]interface{}) string {
	tool, exists := r.GetTool(toolName)
	if !exists {
		return fmt.Sprintf("Tool %s not implemented", toolName)
	}

	result, err := tool.Run(ctx, input)
	if err != nil {
		return fmt.Sprintf("Error running tool %s: %v", toolName, err)
	}

	// 优化工具执行结果的展示格式
	formattedResult := fmt.Sprintf("=== Tool: %s ===\n%s\n=== End of Tool Result ===", toolName, result)
	return formattedResult
}

// 辅助函数：检查文件路径是否在当前目录内
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

// ShellCommandTool 执行shell命令的工具
type ShellCommandTool struct{}

func (t *ShellCommandTool) Name() string {
	return "execute_shell"
}

func (t *ShellCommandTool) Description() string {
	return "Executes a shell command and returns the output"
}

func (t *ShellCommandTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	// 检查是否提供了command参数
	command, ok := input["command"].(string)
	if !ok {
		return "Error: command parameter is required", nil
	}

	// 安全控制：检查危险命令
	// 只保留真正危险的命令，放宽对常用命令的限制
	dangerousCommands := []string{
		"rm -rf", "rmdir -rf", "shutdown", "reboot", "halt", "poweroff",
		"dd", "mkfs", "format", "fdisk", "parted",
		"kill", "killall", "pkill",
		"sudo", "su", "login", "passwd",
		"wget", "curl", "ftp", "scp", "sftp",
	}

	// 检查命令是否包含危险操作
	cmdLower := strings.ToLower(command)
	for _, dangerousCmd := range dangerousCommands {
		// 确保我们匹配的是完整的命令，而不是命令的一部分
		// 例如，"rm" 应该匹配 "rm file" 但不应该匹配 "ls -lrm"
		if strings.Contains(cmdLower, " "+dangerousCmd+" ") || strings.HasPrefix(cmdLower, dangerousCmd+" ") || strings.HasSuffix(cmdLower, " "+dangerousCmd) {
			return fmt.Sprintf("Security Warning: Command '%s' is potentially dangerous and has been blocked.\nFor safety reasons, this operation is not allowed.", command), nil
		}
	}

	// 显示执行命令的提示
	fmt.Printf("Executing command: %s\n", command)

	// 执行shell命令，添加超时机制
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second) // 30秒超时
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, "bash", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error executing command: Command timed out after 30 seconds\nOutput: %s", string(output)), nil
		}
		return fmt.Sprintf("Error executing command: %v\nOutput: %s", err, string(output)), nil
	}

	// 显示命令执行完成的提示
	fmt.Println("Command executed successfully")
	return string(output), nil
}

// ReadFileTool 读取文件的工具
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Reads the content of a file"
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

// WriteFileTool 写入文件的工具
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Writes content to a file"
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

// EditFileTool 编辑文件的工具
type EditFileTool struct{}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Description() string {
	return "Edits a file by replacing a string"
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

// ListSkillsTool 列出技能的工具
type ListSkillsTool struct {
	skillCatalog *skills.SkillCatalog
}

func (t *ListSkillsTool) Name() string {
	return "list_skills"
}

func (t *ListSkillsTool) Description() string {
	return "Lists all available skills"
}

func (t *ListSkillsTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	skills := t.skillCatalog.GetSkills()
	if len(skills) == 0 {
		return "No skills available", nil
	}

	result := "Available skills:\n"
	for name, skill := range skills {
		result += fmt.Sprintf("- %s: %s\n", name, skill.Description)
	}
	return result, nil
}

// ActivateSkillTool 激活技能的工具
type ActivateSkillTool struct {
	skillCatalog *skills.SkillCatalog
}

func (t *ActivateSkillTool) Name() string {
	return "activate_skill"
}

func (t *ActivateSkillTool) Description() string {
	return "Activates a skill and returns its content"
}

func (t *ActivateSkillTool) Run(ctx context.Context, input map[string]interface{}) (string, error) {
	skillName, ok := input["skill_name"].(string)
	if !ok {
		return "Error: skill_name parameter is required", nil
	}

	skillContent, err := t.skillCatalog.ActivateSkill(skillName)
	if err != nil {
		return fmt.Sprintf("Skill '%s' not found", skillName), nil
	}

	return fmt.Sprintf("Skill '%s' activated. Full skill content:\n\n%s", skillName, skillContent), nil
}
