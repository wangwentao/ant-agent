package impl

import (
	"context"
	"fmt"

	"ant-agent/internal/skills"
)

// ListSkillsTool 列出技能的工具
type ListSkillsTool struct {
	skillCatalog *skills.SkillCatalog
}

func NewListSkillsTool(skillCatalog *skills.SkillCatalog) *ListSkillsTool {
	return &ListSkillsTool{skillCatalog: skillCatalog}
}

func (t *ListSkillsTool) Name() string {
	return "list_skills"
}

func (t *ListSkillsTool) Description() string {
	return "Lists all available skills"
}

func (t *ListSkillsTool) Schema() map[string]interface{} {
	return map[string]interface{}{}
}

func (t *ListSkillsTool) Required() []string {
	return []string{}
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
