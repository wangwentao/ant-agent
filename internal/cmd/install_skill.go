package cmd

import (
	"ant-agent/internal/logs"
	"ant-agent/internal/skills"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// InstallSkillCommand 安装技能命令
type InstallSkillCommand struct{}

// NewInstallSkillCommand 创建新的安装技能命令
func NewInstallSkillCommand() *InstallSkillCommand {
	return &InstallSkillCommand{}
}

// Name 返回命令名称
func (c *InstallSkillCommand) Name() string {
	return "install-skill"
}

// Description 返回命令描述
func (c *InstallSkillCommand) Description() string {
	return "Install skills from directory or GitHub"
}

// Execute 执行命令
func (c *InstallSkillCommand) Execute(args []string, skillCatalog *skills.SkillCatalog) error {
	if len(args) < 1 {
		return fmt.Errorf("Please provide a skill directory path or URL.")
	}

	// 解析参数
	skillPath := ""
	skillNames := []string{}
	installAll := false
	i := 0
	for ; i < len(args); i++ {
		if args[i] == "--skill" && i+1 < len(args) {
			skillNames = append(skillNames, args[i+1])
			i++
		} else if args[i] == "--skills" {
			// 收集所有剩余的参数作为技能名称
			remainingArgs := args[i+1:]
			if len(remainingArgs) > 0 && remainingArgs[0] == "all" {
				installAll = true
			} else {
				skillNames = append(skillNames, remainingArgs...)
			}
			break
		} else if skillPath == "" {
			skillPath = args[i]
		}
	}

	if skillPath == "" {
		return fmt.Errorf("Please provide a skill directory path or URL.")
	}

	// 安装技能
	if installAll || len(skillNames) == 0 {
		// 安装所有技能
		return installSkill(skillPath, "", skillCatalog)
	} else {
		// 安装指定的技能
		return installSkills(skillPath, skillNames, skillCatalog)
	}
}

// installSkill 安装技能并验证其结构是否符合规范
func installSkill(skillPath string, skillName string, skillCatalog *skills.SkillCatalog) error {
	// 获取源目录（本地路径或克隆的GitHub仓库）
	sourceDir, err := getSourceDir(skillPath)
	if err != nil {
		return err
	}
	// 如果是临时目录，确保在函数结束时删除
	if isGitHubPath(skillPath) {
		defer os.RemoveAll(sourceDir)
	}

	// 检查是否存在SKILL.md文件
	skillMdPath := filepath.Join(sourceDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err == nil {
		// 单个技能，直接安装
		return installSingleSkill(sourceDir, skillCatalog)
	}

	// 检查是否存在skills目录（多个技能的情况）
	skillsDir := filepath.Join(sourceDir, "skills")
	if info, err := os.Stat(skillsDir); err != nil || !info.IsDir() {
		return fmt.Errorf("SKILL.md file not found in skill directory")
	}

	if skillName != "" {
		// 安装指定技能
		skillDir := filepath.Join(skillsDir, skillName)
		if info, err := os.Stat(skillDir); err != nil || !info.IsDir() {
			return fmt.Errorf("skill %s not found in skills directory", skillName)
		}
		return installSingleSkill(skillDir, skillCatalog)
	}

	// 安装所有技能
	return installMultipleSkills(skillsDir, []string{}, skillCatalog)
}

// installSkills 安装指定的技能
func installSkills(skillPath string, skillNames []string, skillCatalog *skills.SkillCatalog) error {
	// 获取源目录（本地路径或克隆的GitHub仓库）
	sourceDir, err := getSourceDir(skillPath)
	if err != nil {
		return err
	}
	// 如果是临时目录，确保在函数结束时删除
	if isGitHubPath(skillPath) {
		defer os.RemoveAll(sourceDir)
	}

	// 检查是否存在SKILL.md文件
	skillMdPath := filepath.Join(sourceDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err == nil {
		// 单个技能，直接安装
		return installSingleSkill(sourceDir, skillCatalog)
	}

	// 检查是否存在skills目录（多个技能的情况）
	skillsDir := filepath.Join(sourceDir, "skills")
	if info, err := os.Stat(skillsDir); err != nil || !info.IsDir() {
		return fmt.Errorf("skills directory not found in the repository")
	}

	// 安装指定的技能
	return installMultipleSkills(skillsDir, skillNames, skillCatalog)
}

// getSourceDir 获取技能源目录
// 如果是GitHub路径，克隆到临时目录；否则返回本地路径
func getSourceDir(skillPath string) (string, error) {
	if isGitHubPath(skillPath) {
		// 克隆GitHub仓库到临时目录
		tempDir, err := cloneGitHubRepo(skillPath)
		if err != nil {
			return "", fmt.Errorf("failed to clone GitHub repo: %w", err)
		}
		return tempDir, nil
	}

	// 检查本地路径是否存在
	info, err := os.Stat(skillPath)
	if err != nil {
		return "", fmt.Errorf("skill path does not exist: %w", err)
	}

	// 确保是目录
	if !info.IsDir() {
		return "", fmt.Errorf("skill path must be a directory")
	}

	return skillPath, nil
}

// installSingleSkill 安装单个技能
func installSingleSkill(sourceDir string, skillCatalog *skills.SkillCatalog) error {
	// 检查是否存在SKILL.md文件
	skillMdPath := filepath.Join(sourceDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err != nil {
		return fmt.Errorf("SKILL.md file not found in skill directory")
	}

	// 解析SKILL.md文件，验证结构
	skill, err := skills.ParseSkill(skillMdPath)
	if err != nil {
		return fmt.Errorf("invalid SKILL.md file: %w", err)
	}

	// 验证技能名称和描述
	if skill.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if skill.Description == "" {
		return fmt.Errorf("skill description is required")
	}

	// 确定目标安装目录
	targetDir := filepath.Join(".ant", "skills", skill.Name)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 复制技能文件
	if err := copySkillFiles(sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy skill files: %w", err)
	}

	// 重新发现技能
	if err := skillCatalog.DiscoverSkills(); err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	return nil
}

// installMultipleSkills 安装多个技能
// 如果skillNames为空，安装skills目录中的所有技能
// 否则，只安装指定名称的技能
func installMultipleSkills(skillsDir string, skillNames []string, skillCatalog *skills.SkillCatalog) error {
	if len(skillNames) == 0 {
		// 安装所有技能
		return installAllSkills(skillsDir, skillCatalog)
	}

	// 安装指定的技能
	return installSpecifiedSkills(skillsDir, skillNames, skillCatalog)
}

// installAllSkills 安装skills目录中的所有技能
func installAllSkills(skillsDir string, skillCatalog *skills.SkillCatalog) error {
	// 遍历skills目录中的所有子目录
	subDirs, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	installedCount := 0

	for _, subDir := range subDirs {
		if subDir.IsDir() {
			skillDir := filepath.Join(skillsDir, subDir.Name())
			// 检查是否存在SKILL.md文件
			skillMdPath := filepath.Join(skillDir, "SKILL.md")
			if _, err := os.Stat(skillMdPath); err == nil {
				// 安装单个技能
				if err := installSingleSkill(skillDir, skillCatalog); err != nil {
					logs.Warn("Failed to install skill %s: %v", subDir.Name(), err)
					continue
				}
				installedCount++
			}
		}
	}

	if installedCount == 0 {
		return fmt.Errorf("no skills found in skills directory")
	}

	logs.Info("Successfully installed %d skills", installedCount)
	return nil
}

// installSpecifiedSkills 安装指定名称的技能
func installSpecifiedSkills(skillsDir string, skillNames []string, skillCatalog *skills.SkillCatalog) error {
	// 安装指定的技能
	installedCount := 0
	for _, skillName := range skillNames {
		skillDir := filepath.Join(skillsDir, skillName)
		if info, err := os.Stat(skillDir); err == nil && info.IsDir() {
			if err := installSingleSkill(skillDir, skillCatalog); err != nil {
				logs.Warn("Failed to install skill %s: %v", skillName, err)
				continue
			}
			installedCount++
		} else {
			logs.Warn("Skill %s not found in skills directory", skillName)
		}
	}

	if installedCount == 0 {
		return fmt.Errorf("no skills installed")
	}

	logs.Info("Successfully installed %d skills", installedCount)
	return nil
}

// isGitHubPath 检查是否为GitHub路径
func isGitHubPath(path string) bool {
	// 检查是否为GitHub shorthand (owner/repo)
	ghShorthand := regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`)
	if ghShorthand.MatchString(path) {
		return true
	}

	// 检查是否为GitHub URL
	ghURL := regexp.MustCompile(`^https?://github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+`)
	if ghURL.MatchString(path) {
		return true
	}

	// 检查是否为Git URL
	gitURL := regexp.MustCompile(`^git@github\.com:[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+\.git$`)
	if gitURL.MatchString(path) {
		return true
	}

	return false
}

// cloneGitHubRepo 克隆GitHub仓库到临时目录
func cloneGitHubRepo(repoPath string) (string, error) {
	// 生成临时目录
	tempDir, err := os.MkdirTemp("", "ant-skill-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 构建git clone命令
	var repoURL string

	// 处理不同格式的GitHub路径
	if strings.Contains(repoPath, "github.com") {
		// 已经是完整URL
		repoURL = repoPath
	} else if strings.Contains(repoPath, ":") {
		// Git URL format
		repoURL = repoPath
	} else {
		// GitHub shorthand format (owner/repo)
		repoURL = "https://github.com/" + repoPath
	}

	// 执行git clone命令
	cmd := exec.Command("git", "clone", repoURL, tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	return tempDir, nil
}

// copySkillFiles 复制技能文件到目标目录
func copySkillFiles(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 跳过隐藏文件和目录（除了源目录本身）
		if strings.HasPrefix(relPath, ".") && relPath != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 构建目标路径
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// 创建目录
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// 复制文件
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// 设置文件权限
			return os.Chmod(dstPath, info.Mode())
		}
	})
}
