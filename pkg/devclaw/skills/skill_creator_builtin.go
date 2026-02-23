package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type skillCreatorSkill struct {
	skillsDir string
	registry  *Registry
	installer *Installer
}

func newSkillCreatorSkill(skillsDir string, registry *Registry, installer *Installer) Skill {
	return &skillCreatorSkill{
		skillsDir: skillsDir,
		registry:  registry,
		installer: installer,
	}
}

func (s *skillCreatorSkill) Metadata() Metadata {
	return Metadata{
		Name:        "skill-creator",
		Description: "Design, create, and manage DevClaw skills using Go-native tools. Follows OpenClaw engineering standards.",
		Version:     "2.0.0",
		Author:      "DevClaw Core",
	}
}

func (s *skillCreatorSkill) Tools() []Tool {
	return []Tool{
		{
			Name:        "init_skill",
			Description: "Create a new skill directory with a standard SKILL.md template.",
			Parameters: []ToolParameter{
				{Name: "name", Type: "string", Description: "Skill name (kebab-case)", Required: true},
				{Name: "description", Type: "string", Description: "Concise description for triggering", Required: true},
				{Name: "instructions", Type: "string", Description: "Initial instructions (Markdown)", Required: false},
			},
			Handler: s.handleInitSkill,
		},
		{
			Name:        "add_script",
			Description: "Add a script (Python, JS, etc.) to a skill's scripts/ directory.",
			Parameters: []ToolParameter{
				{Name: "skill_name", Type: "string", Description: "Target skill name", Required: true},
				{Name: "script_name", Type: "string", Description: "Filename (e.g. tool.py)", Required: true},
				{Name: "content", Type: "string", Description: "Script source code", Required: true},
			},
			Handler: s.handleAddScript,
		},
		{
			Name:        "list_skills",
			Description: "List all loaded and available user skills.",
			Handler:     s.handleListSkills,
		},
		{
			Name:        "test_skill",
			Description: "Execute a skill with test input to verify behavior.",
			Parameters: []ToolParameter{
				{Name: "name", Type: "string", Description: "Skill name", Required: true},
				{Name: "input", Type: "string", Description: "Test input string", Required: true},
			},
			Handler: s.handleTestSkill,
		},
		{
			Name:        "install_skill",
			Description: "Install a skill from ClawHub or a URL.",
			Parameters: []ToolParameter{
				{Name: "source", Type: "string", Description: "Slug or URL", Required: true},
			},
			Handler: s.handleInstallSkill,
		},
	}
}

func (s *skillCreatorSkill) SystemPrompt() string {
	return `You are an expert Skill Engineer. Follow these principles when creating or updating skills:

### 1. Progressive Disclosure
- Metadata (Description): Should be extremely concise for routing.
- SKILL.md: Core logic and instructions. Keep under 500 lines.
- Resources (scripts/references): For heavy lifting. Link to them from SKILL.md.

### 2. Design Methodology (OODA)
- **Observe**: Analyze the user's domain and repetitive tasks.
- **Orient**: Choose between high freedom (Prompt) and low freedom (Script).
- **Decide**: Plan the file structure (SKILL.md, scripts/, references/).
- **Act**: Use 'init_skill' to scaffold and 'add_script' to implement.

### 3. Skill Anatomy
- Always kebab-case for names.
- Descriptions must start with "Use when..." or "Handles...".
- Prefer declarative examples over verbose explanations.`
}

func (s *skillCreatorSkill) Triggers() []string { return nil }
func (s *skillCreatorSkill) Init(ctx context.Context, config map[string]any) error {
	return nil
}
func (s *skillCreatorSkill) Execute(ctx context.Context, input string) (string, error) {
	return "Follow the system prompt instructions to manage skills using the provided tools.", nil
}
func (s *skillCreatorSkill) Shutdown() error { return nil }

// --- Handlers (Adapted from copilot/skill_creator.go) ---

func (s *skillCreatorSkill) handleInitSkill(ctx context.Context, args map[string]any) (any, error) {
	name, _ := args["name"].(string)
	desc, _ := args["description"].(string)
	instr, _ := args["instructions"].(string)

	name = s.sanitizeName(name)
	dir := filepath.Join(s.skillsDir, name)
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		return nil, err
	}

	if instr == "" {
		instr = fmt.Sprintf("# %s\n\nInstructions go here.", strings.Title(strings.ReplaceAll(name, "-", " ")))
	}

	content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n%s", name, desc, instr)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		return nil, err
	}

	return fmt.Sprintf("Skill '%s' initialized at %s", name, dir), nil
}

func (s *skillCreatorSkill) handleAddScript(ctx context.Context, args map[string]any) (any, error) {
	skillName, _ := args["skill_name"].(string)
	scriptName, _ := args["script_name"].(string)
	content, _ := args["content"].(string)

	path := filepath.Join(s.skillsDir, s.sanitizeName(skillName), "scripts", scriptName)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return nil, err
	}
	return fmt.Sprintf("Script '%s' added to skill '%s'.", scriptName, skillName), nil
}

func (s *skillCreatorSkill) handleListSkills(ctx context.Context, args map[string]any) (any, error) {
	skills := s.registry.List()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Loaded skills (%d):\n", len(skills)))
	for _, m := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", m.Name, m.Description))
	}
	return sb.String(), nil
}

func (s *skillCreatorSkill) handleTestSkill(ctx context.Context, args map[string]any) (any, error) {
	name, _ := args["name"].(string)
	input, _ := args["input"].(string)

	sk, ok := s.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	// Use a 30s timeout for tests
	tCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return sk.Execute(tCtx, input)
}

func (s *skillCreatorSkill) handleInstallSkill(ctx context.Context, args map[string]any) (any, error) {
	if s.installer == nil {
		return nil, fmt.Errorf("installer not configured")
	}
	source, _ := args["source"].(string)
	res, err := s.installer.Install(ctx, source)
	if err != nil {
		return nil, err
	}

	// Try hot-reload
	s.registry.Reload(ctx)

	return fmt.Sprintf("Skill '%s' installed to %s", res.Name, res.Path), nil
}

func (s *skillCreatorSkill) sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	var clean strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			clean.WriteRune(r)
		}
	}
	return clean.String()
}
