package skills

import (
	"os"
	"path/filepath"
	"strings"

	"ant-agent/internal/logs"
)

// Skill represents a single skill
type Skill struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Location      string            `json:"location"` // Path to SKILL.md
	License       string            `json:"license,omitempty"`
	Compatibility string            `json:"compatibility,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedTools  []string          `json:"allowed_tools,omitempty"`
	Body          string            `json:"body"` // Markdown content after frontmatter
}

// SkillCatalog manages all available skills
type SkillCatalog struct {
	skills map[string]*Skill
}

// NewSkillCatalog creates a new skill catalog
func NewSkillCatalog() *SkillCatalog {
	return &SkillCatalog{
		skills: make(map[string]*Skill),
	}
}

// GetSkills returns all available skills
func (c *SkillCatalog) GetSkills() map[string]*Skill {
	return c.skills
}

// DiscoverSkills discovers skills from the filesystem
func (c *SkillCatalog) DiscoverSkills() error {
	// Clear existing skills before re-scanning
	c.skills = make(map[string]*Skill)

	// Scan common skill directories
	dirsToScan := []string{
		// Project-level
		"./.agents/skills",
		// User-level (cross-platform)
		filepath.Join(os.Getenv("HOME"), ".agents/skills"),
	}

	logs.Debug("Scanning for skills in directories:")
	for _, dir := range dirsToScan {
		logs.Debug("  - %s", dir)
		if err := c.scanDirectory(dir); err != nil {
			// Continue scanning other directories if one fails
			logs.Warn("Failed to scan directory %s: %v", dir, err)
			continue
		}
	}

	logs.Debug("Skill discovery complete. Found %d skills", len(c.skills))
	for name := range c.skills {
		logs.Debug("  - %s", name)
	}

	return nil
}

// scanDirectory scans a directory for skills
func (c *SkillCatalog) scanDirectory(dir string) error {
	logs.Debug("Walking directory: %s", dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logs.Warn("Error walking path %s: %v", path, err)
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			logs.Debug("Skipping hidden directory: %s", path)
			return filepath.SkipDir
		}

		// Only process directories (SKILL.md files are in directories)
		if !info.IsDir() {
			return nil
		}

		// Check if this directory contains a SKILL.md file
		skillMdPath := filepath.Join(path, "SKILL.md")
		logs.Debug("Checking for SKILL.md in: %s", path)
		if _, err := os.Stat(skillMdPath); err == nil {
			// Found a skill, parse it
			logs.Debug("Found SKILL.md at: %s", skillMdPath)
			skill, err := ParseSkill(skillMdPath)
			if err != nil {
				// Skip invalid skills
				logs.Warn("Failed to parse skill at %s: %v", skillMdPath, err)
				return nil
			}

			// Add to catalog, project-level skills override user-level
			logs.Debug("Adding skill to catalog: %s", skill.Name)
			c.skills[skill.Name] = skill
		}

		return nil
	})
}

// GetSkill returns a skill by name
func (c *SkillCatalog) GetSkill(name string) (*Skill, bool) {
	skill, exists := c.skills[name]
	return skill, exists
}

// ActivateSkill activates a skill and returns its content
func (c *SkillCatalog) ActivateSkill(name string) (string, error) {
	skill, exists := c.GetSkill(name)
	if !exists {
		return "", os.ErrNotExist
	}

	// Return the skill content with structured wrapping
	return formatSkillContent(skill), nil
}

// formatSkillContent formats skill content for the model
func formatSkillContent(skill *Skill) string {
	var sb strings.Builder

	// Add structured wrapper
	sb.WriteString("<skill_content name=\"" + skill.Name + "\">\n")
	sb.WriteString(skill.Body)
	sb.WriteString("\n\nSkill directory: " + filepath.Dir(skill.Location) + "\n")
	sb.WriteString("Relative paths in this skill are relative to the skill directory.\n\n")

	// List bundled resources
	sb.WriteString("<skill_resources>\n")
	dir := filepath.Dir(skill.Location)
	for _, subdir := range []string{"scripts", "references", "assets"} {
		subdirPath := filepath.Join(dir, subdir)
		if info, err := os.Stat(subdirPath); err == nil && info.IsDir() {
			files, _ := os.ReadDir(subdirPath)
			for _, file := range files {
				if !file.IsDir() {
					sb.WriteString("  <file>" + subdir + "/" + file.Name() + "</file>\n")
				}
			}
		}
	}
	sb.WriteString("</skill_resources>\n")
	sb.WriteString("</skill_content>")

	return sb.String()
}
