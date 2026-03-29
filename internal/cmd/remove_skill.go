package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"ant-agent/internal/skills"
)

// RemoveSkillCommand 删除技能命令
type RemoveSkillCommand struct {}

// NewRemoveSkillCommand 创建新的删除技能命令
func NewRemoveSkillCommand() *RemoveSkillCommand {
	return &RemoveSkillCommand{}
}

// Name 返回命令名称
func (c *RemoveSkillCommand) Name() string {
	return "remove-skill"
}

// Description 返回命令描述
func (c *RemoveSkillCommand) Description() string {
	return "Remove a skill by name"
}

// Execute 执行命令
func (c *RemoveSkillCommand) Execute(args []string, skillCatalog *skills.SkillCatalog) error {
	if len(args) < 1 {
		return fmt.Errorf("Please provide a skill name.")
	}

	skillName := args[0]
	return removeSkill(skillName, skillCatalog)
}

// removeSkill 删除指定名称的技能
func removeSkill(skillName string, skillCatalog *skills.SkillCatalog) error {
	// 检查技能是否存在
	skill, exists := skillCatalog.GetSkill(skillName)
	if !exists {
		return fmt.Errorf("skill '%s' not found", skillName)
	}

	// 确定技能目录路径
	skillDir := filepath.Dir(skill.Location)

	// 删除技能目录
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	// 重新发现技能，更新技能目录
	if err := skillCatalog.DiscoverSkills(); err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	return nil
}
