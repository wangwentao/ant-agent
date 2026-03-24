package tools_test

import (
	"context"
	"testing"

	"ant-agent/internal/skills"
	"ant-agent/internal/tools"
)

func TestToolRegistry(t *testing.T) {
	// 创建技能目录
	skillCatalog := skills.NewSkillCatalog()

	// 创建工具注册表
	registry := tools.NewToolRegistry(skillCatalog)

	// 测试获取所有工具
	allTools := registry.GetAllTools()
	if len(allTools) == 0 {
		t.Error("Expected at least one tool, got none")
	}

	// 测试获取特定工具
	shellTool, exists := registry.GetTool("execute_shell")
	if !exists {
		t.Error("Expected execute_shell tool to exist")
	}
	if shellTool.Name() != "execute_shell" {
		t.Errorf("Expected tool name to be 'execute_shell', got %s", shellTool.Name())
	}

	readFileTool, exists := registry.GetTool("read_file")
	if !exists {
		t.Error("Expected read_file tool to exist")
	}
	if readFileTool.Name() != "read_file" {
		t.Errorf("Expected tool name to be 'read_file', got %s", readFileTool.Name())
	}

	writeFileTool, exists := registry.GetTool("write_file")
	if !exists {
		t.Error("Expected write_file tool to exist")
	}
	if writeFileTool.Name() != "write_file" {
		t.Errorf("Expected tool name to be 'write_file', got %s", writeFileTool.Name())
	}

	editFileTool, exists := registry.GetTool("edit_file")
	if !exists {
		t.Error("Expected edit_file tool to exist")
	}
	if editFileTool.Name() != "edit_file" {
		t.Errorf("Expected tool name to be 'edit_file', got %s", editFileTool.Name())
	}

	listSkillsTool, exists := registry.GetTool("list_skills")
	if !exists {
		t.Error("Expected list_skills tool to exist")
	}
	if listSkillsTool.Name() != "list_skills" {
		t.Errorf("Expected tool name to be 'list_skills', got %s", listSkillsTool.Name())
	}

	activateSkillTool, exists := registry.GetTool("activate_skill")
	if !exists {
		t.Error("Expected activate_skill tool to exist")
	}
	if activateSkillTool.Name() != "activate_skill" {
		t.Errorf("Expected tool name to be 'activate_skill', got %s", activateSkillTool.Name())
	}
}

func TestRunTool(t *testing.T) {
	// 创建技能目录
	skillCatalog := skills.NewSkillCatalog()

	// 创建工具注册表
	registry := tools.NewToolRegistry(skillCatalog)

	// 测试运行list_skills工具
	ctx := context.Background()
	input := map[string]interface{}{}
	result := registry.RunTool(ctx, "list_skills", input)
	if result == "" {
		t.Error("Expected result from list_skills tool, got empty string")
	}

	// 测试运行不存在的工具
	result = registry.RunTool(ctx, "nonexistent_tool", input)
	if result == "" {
		t.Error("Expected error message for nonexistent tool, got empty string")
	}
}
