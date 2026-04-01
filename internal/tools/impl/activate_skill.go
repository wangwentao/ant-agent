package impl

import (
	"context"
	"fmt"

	"ant-agent/internal/skills"
)

// ActivateSkillTool 激活技能的工具
type ActivateSkillTool struct {
	skillCatalog *skills.SkillCatalog
}

func NewActivateSkillTool(skillCatalog *skills.SkillCatalog) *ActivateSkillTool {
	return &ActivateSkillTool{skillCatalog: skillCatalog}
}

func (t *ActivateSkillTool) Name() string {
	return "activate_skill"
}

func (t *ActivateSkillTool) Description() string {
	return "Activates a skill and returns its content"
}

func (t *ActivateSkillTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"skill_name": map[string]interface{}{
			"type":        "string",
			"description": "The name of the skill to activate",
		},
	}
}

func (t *ActivateSkillTool) Required() []string {
	return []string{"skill_name"}
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
