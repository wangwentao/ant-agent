package skills

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// 自定义错误类型
var (
	ErrInvalidSkill = errors.New("invalid skill")
	ErrMissingField = errors.New("missing required field")
	ErrInvalidYAML = errors.New("invalid YAML")
)

// ParseSkill parses a SKILL.md file
func ParseSkill(skillMdPath string) (*Skill, error) {
	// Read the file
	file, err := os.Open(skillMdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open skill file: %w", err)
	}
	defer file.Close()

	// Read content
	scanner := bufio.NewScanner(file)
	var content []string
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	// Find frontmatter boundaries
	var frontmatter []string
	var body []string
	inFrontmatter := false
	frontmatterEnded := false

	for _, line := range content {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				inFrontmatter = false
				frontmatterEnded = true
			}
			continue
		}

		if inFrontmatter {
			frontmatter = append(frontmatter, line)
		} else if frontmatterEnded {
			body = append(body, line)
		}
	}

	// Parse frontmatter
	var skill Skill
	if len(frontmatter) > 0 {
		frontmatterYAML := strings.Join(frontmatter, "\n")
		if err := yaml.Unmarshal([]byte(frontmatterYAML), &skill); err != nil {
			// Try to handle malformed YAML gracefully
			if err := parseLenientYAML(frontmatterYAML, &skill); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidYAML, err.Error())
			}
		}
	}

	// Set required fields
	skill.Location = skillMdPath

	// Set name from directory if not provided
	if skill.Name == "" {
		skill.Name = filepath.Base(filepath.Dir(skillMdPath))
	}

	// Set body
	skill.Body = strings.Join(body, "\n")

	// Validate skill
	if err := validateSkill(&skill); err != nil {
		return nil, err
	}

	return &skill, nil
}

// validateSkill validates a skill
func validateSkill(skill *Skill) error {
	// Validate required fields
	if skill.Name == "" {
		return fmt.Errorf("%w: name", ErrMissingField)
	}
	if skill.Description == "" {
		return fmt.Errorf("%w: description", ErrMissingField)
	}

	// Validate skill name (no spaces, only alphanumeric, underscore, hyphen)
	for _, char := range skill.Name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return fmt.Errorf("%w: skill name can only contain alphanumeric characters, underscore, and hyphen", ErrInvalidSkill)
		}
	}

	return nil
}

// parseLenientYAML parses YAML with lenient handling for common issues
func parseLenientYAML(yamlStr string, skill *Skill) error {
	// Simple parser for common cases
	lines := strings.Split(yamlStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes from value if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		case "license":
			skill.License = value
		case "compatibility":
			skill.Compatibility = value
		case "allowed-tools":
			skill.AllowedTools = strings.Fields(value)
		case "version":
			if skill.Metadata == nil {
				skill.Metadata = make(map[string]string)
			}
			skill.Metadata["version"] = value
		case "author":
			if skill.Metadata == nil {
				skill.Metadata = make(map[string]string)
			}
			skill.Metadata["author"] = value
		case "tags":
			if skill.Metadata == nil {
				skill.Metadata = make(map[string]string)
			}
			skill.Metadata["tags"] = value
		}
	}

	return nil
}
