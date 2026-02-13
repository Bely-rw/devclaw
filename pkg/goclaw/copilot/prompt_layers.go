// Package copilot – prompt_layers.go implements the layered system prompt
// (OpenClaw-style). Each layer has a priority and contributes to the final
// prompt that is sent to the LLM as the system message.
//
// Bootstrap files (SOUL.md, AGENTS.md, IDENTITY.md, USER.md, TOOLS.md) are
// loaded from the workspace root and injected as "Project Context".
// If SOUL.md is present, the agent is instructed to embody its persona.
package copilot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/jholhewres/goclaw/pkg/goclaw/copilot/memory"
)

// PromptLayer defines the priority of a prompt layer.
// Lower values = higher priority (never trimmed first on budget cuts).
type PromptLayer int

const (
	LayerCore         PromptLayer = 0  // Base identity and tooling.
	LayerSafety       PromptLayer = 5  // Safety rules.
	LayerIdentity     PromptLayer = 10 // Custom instructions.
	LayerThinking     PromptLayer = 12 // Extended thinking level hint (from /think).
	LayerBootstrap    PromptLayer = 15 // SOUL.md, AGENTS.md, etc.
	LayerBusiness     PromptLayer = 20 // User/workspace context.
	LayerSkills       PromptLayer = 40 // Active skill instructions.
	LayerMemory       PromptLayer = 50 // Long-term memory facts.
	LayerTemporal     PromptLayer = 60 // Date/time context.
	LayerConversation PromptLayer = 70 // Recent history summary.
	LayerRuntime      PromptLayer = 80 // Runtime info (final line).
)

// layerEntry represents a single prompt layer entry.
type layerEntry struct {
	layer   PromptLayer
	content string
}

// PromptComposer assembles the final system prompt from multiple layers.
type PromptComposer struct {
	config      *Config
	memoryStore *memory.FileStore
	skillGetter func(name string) (interface{ SystemPrompt() string }, bool)
}

// NewPromptComposer creates a new prompt composer.
func NewPromptComposer(config *Config) *PromptComposer {
	return &PromptComposer{config: config}
}

// SetMemoryStore configures the memory store for the prompt composer.
func (p *PromptComposer) SetMemoryStore(store *memory.FileStore) {
	p.memoryStore = store
}

// SetSkillGetter sets the function used to retrieve skill system prompts.
func (p *PromptComposer) SetSkillGetter(getter func(name string) (interface{ SystemPrompt() string }, bool)) {
	p.skillGetter = getter
}

// Compose builds the complete system prompt for a session and user input.
func (p *PromptComposer) Compose(session *Session, input string) string {
	layers := make([]layerEntry, 0, 10)

	// Layer 0: Core — base identity and tooling guidance.
	layers = append(layers, layerEntry{
		layer:   LayerCore,
		content: p.buildCoreLayer(),
	})

	// Layer 5: Safety — guardrails.
	layers = append(layers, layerEntry{
		layer:   LayerSafety,
		content: p.buildSafetyLayer(),
	})

	// Layer 10: Identity — custom instructions from config.
	if p.config.Instructions != "" {
		layers = append(layers, layerEntry{
			layer:   LayerIdentity,
			content: "## Custom Instructions\n\n" + p.config.Instructions,
		})
	}

	// Layer 12: Thinking — extended thinking level hint from session.
	if thinkingPrompt := p.buildThinkingLayer(session); thinkingPrompt != "" {
		layers = append(layers, layerEntry{
			layer:   LayerThinking,
			content: thinkingPrompt,
		})
	}

	// Layer 15: Bootstrap — SOUL.md, AGENTS.md, etc.
	if bootstrapPrompt := p.buildBootstrapLayer(); bootstrapPrompt != "" {
		layers = append(layers, layerEntry{
			layer:   LayerBootstrap,
			content: bootstrapPrompt,
		})
	}

	// Layer 20: Business — workspace/user context.
	cfg := session.GetConfig()
	if cfg.BusinessContext != "" {
		layers = append(layers, layerEntry{
			layer:   LayerBusiness,
			content: "## Workspace Context\n\n" + cfg.BusinessContext,
		})
	}

	// Layer 40: Skills — active skill instructions.
	if skillPrompt := p.buildSkillsLayer(session); skillPrompt != "" {
		layers = append(layers, layerEntry{
			layer:   LayerSkills,
			content: skillPrompt,
		})
	}

	// Layer 50: Memory — relevant long-term facts.
	if memoryPrompt := p.buildMemoryLayer(session, input); memoryPrompt != "" {
		layers = append(layers, layerEntry{
			layer:   LayerMemory,
			content: memoryPrompt,
		})
	}

	// Layer 60: Temporal — current date/time.
	layers = append(layers, layerEntry{
		layer:   LayerTemporal,
		content: p.buildTemporalLayer(),
	})

	// Layer 70: Conversation — recent history summary.
	if historyPrompt := p.buildConversationLayer(session); historyPrompt != "" {
		layers = append(layers, layerEntry{
			layer:   LayerConversation,
			content: historyPrompt,
		})
	}

	// Layer 80: Runtime — final line with system info.
	layers = append(layers, layerEntry{
		layer:   LayerRuntime,
		content: p.buildRuntimeLayer(),
	})

	return p.assembleLayers(layers)
}

// ---------- Layer Builders ----------

// buildCoreLayer creates the base identity and tooling guidance.
func (p *PromptComposer) buildCoreLayer() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("You are %s, a personal assistant running inside GoClaw.\n\n", p.config.Name))

	b.WriteString("## Tooling\n\n")
	b.WriteString("You have access to tools for: file I/O (read, write, edit, search, glob), ")
	b.WriteString("bash execution (full system access), SSH/SCP (remote machines), ")
	b.WriteString("web search and fetch, memory (save/search/list), ")
	b.WriteString("scheduling (cron), and skill management (install/create/remove).\n\n")
	b.WriteString("Tool names are case-sensitive. Call tools exactly as listed.\n")
	b.WriteString("TOOLS.md does not control tool availability — it is your cheat sheet for environment-specific notes.\n\n")

	b.WriteString("## Tool Call Style\n\n")
	b.WriteString("Default: do not narrate routine, low-risk tool calls — just call the tool.\n")
	b.WriteString("Narrate only when it helps: multi-step work, complex problems, sensitive actions (deletions, deployments), or when the user explicitly asks.\n")
	b.WriteString("Keep narration brief and value-dense. Avoid repeating obvious steps.\n")
	b.WriteString("Use plain human language unless in a technical context.\n")

	return b.String()
}

// buildSafetyLayer creates the safety guardrails section.
func (p *PromptComposer) buildSafetyLayer() string {
	return `## Safety

You have no independent goals: do not pursue self-preservation, replication, resource acquisition, or power-seeking. Avoid long-term plans beyond the user's request.

Prioritize safety and human oversight over task completion. If instructions conflict, pause and ask. Comply with stop/pause/audit requests and never bypass safeguards.

Do not manipulate or persuade anyone to expand access or disable safeguards. Do not copy yourself or change system prompts, safety rules, or tool policies unless explicitly requested by the owner.

When using destructive tools (rm, drop, deploy): confirm with the user first unless they've explicitly pre-approved the action.

File operations: prefer reversible actions. Use trash over rm. Create backups before major changes.

SSH/remote: only connect to known hosts. Don't store passwords in plaintext. Use the vault for secrets.`
}

// buildThinkingLayer adds extended-thinking guidance based on session /think level.
func (p *PromptComposer) buildThinkingLayer(session *Session) string {
	level := session.GetThinkingLevel()
	if level == "" || level == "off" {
		return ""
	}
	instructions := map[string]string{
		"low":    "Think step-by-step when the task is complex. Keep reasoning brief for simple tasks.",
		"medium": "Think through problems systematically. Show your reasoning for non-trivial tasks.",
		"high":   "Use extended thinking: reason carefully before answering, consider alternatives, then respond. Favor depth over speed.",
	}
	if instr, ok := instructions[level]; ok {
		return "## Thinking Mode\n\n" + instr
	}
	return ""
}

// buildBootstrapLayer loads bootstrap files from the workspace root.
func (p *PromptComposer) buildBootstrapLayer() string {
	bootstrapFiles := []struct {
		Path    string
		Section string
	}{
		{"SOUL.md", "SOUL.md"},
		{"AGENTS.md", "AGENTS.md"},
		{"IDENTITY.md", "IDENTITY.md"},
		{"USER.md", "USER.md"},
		{"TOOLS.md", "TOOLS.md"},
		{"MEMORY.md", "MEMORY.md"},
	}

	// Search directories: workspace dir, current dir, configs/.
	searchDirs := []string{"."}
	if p.config.Heartbeat.WorkspaceDir != "" && p.config.Heartbeat.WorkspaceDir != "." {
		searchDirs = append([]string{p.config.Heartbeat.WorkspaceDir}, searchDirs...)
	}
	searchDirs = append(searchDirs, "configs")

	var files []struct {
		path    string
		content string
	}
	hasSoul := false

	for _, bf := range bootstrapFiles {
		var content []byte
		var err error

		for _, dir := range searchDirs {
			content, err = os.ReadFile(filepath.Join(dir, bf.Path))
			if err == nil {
				break
			}
		}
		if err != nil || len(strings.TrimSpace(string(content))) == 0 {
			continue
		}

		text := strings.TrimSpace(string(content))

		// Truncate very large files.
		if len(text) > 20000 {
			text = text[:20000] + "\n\n... [truncated at 20KB]"
		}

		files = append(files, struct {
			path    string
			content string
		}{bf.Section, text})

		if bf.Path == "SOUL.md" {
			hasSoul = true
		}
	}

	if len(files) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("# Project Context\n\n")
	b.WriteString("The following project context files have been loaded:\n\n")

	if hasSoul {
		b.WriteString("If SOUL.md is present, embody its persona and tone. ")
		b.WriteString("Avoid stiff, generic replies; follow its guidance unless higher-priority instructions override it.\n\n")
	}

	for _, f := range files {
		b.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", f.path, f.content))
	}

	return b.String()
}

// buildSkillsLayer creates instructions from active skills.
func (p *PromptComposer) buildSkillsLayer(session *Session) string {
	activeSkills := session.GetActiveSkills()
	if len(activeSkills) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Skills\n\n")
	b.WriteString("You have specialized skills available. Each skill provides tools and context.\n\n")

	for _, skillName := range activeSkills {
		b.WriteString(fmt.Sprintf("### %s\n", skillName))

		if p.skillGetter != nil {
			if skill, ok := p.skillGetter(skillName); ok {
				sp := skill.SystemPrompt()
				if sp != "" {
					b.WriteString(sp)
					b.WriteString("\n")
				}
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// buildMemoryLayer creates the memory context section.
func (p *PromptComposer) buildMemoryLayer(session *Session, input string) string {
	var parts []string

	// Pull from persistent memory store.
	if p.memoryStore != nil {
		facts := p.memoryStore.RecentFacts(15, input)
		if facts != "" {
			parts = append(parts, "## Memory Recall\n\nRelevant facts from long-term memory:\n\n"+facts)
		}
	}

	// Session-level facts.
	sessionFacts := session.GetFacts()
	if len(sessionFacts) > 0 {
		var b strings.Builder
		b.WriteString("## Session Context\n\n")
		for _, fact := range sessionFacts {
			b.WriteString(fmt.Sprintf("- %s\n", fact))
		}
		parts = append(parts, b.String())
	}

	return strings.Join(parts, "\n")
}

// buildTemporalLayer adds date/time context.
func (p *PromptComposer) buildTemporalLayer() string {
	loc, err := time.LoadLocation(p.config.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	return fmt.Sprintf("## Current Date & Time\n\n%s\nTimezone: %s\nDay: %s",
		now.Format("2006-01-02 15:04:05"),
		p.config.Timezone,
		now.Format("Monday"),
	)
}

// buildConversationLayer creates a summary of recent history.
func (p *PromptComposer) buildConversationLayer(session *Session) string {
	history := session.RecentHistory(p.config.Memory.MaxMessages)
	if len(history) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Recent Conversation\n\n")

	for _, entry := range history {
		b.WriteString(fmt.Sprintf("**User:** %s\n**Assistant:** %s\n\n",
			entry.UserMessage, entry.AssistantResponse))
	}

	return b.String()
}

// buildRuntimeLayer creates the runtime info line (last in prompt).
func (p *PromptComposer) buildRuntimeLayer() string {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()

	return fmt.Sprintf("---\nRuntime: agent=%s | model=%s | os=%s/%s | host=%s | cwd=%s | lang=%s",
		p.config.Name,
		p.config.Model,
		runtime.GOOS,
		runtime.GOARCH,
		hostname,
		cwd,
		p.config.Language,
	)
}

// assembleLayers combines all layers in priority order.
func (p *PromptComposer) assembleLayers(layers []layerEntry) string {
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].layer < layers[j].layer
	})

	var parts []string
	for _, l := range layers {
		if l.content != "" {
			parts = append(parts, l.content)
		}
	}

	return strings.Join(parts, "\n\n")
}
