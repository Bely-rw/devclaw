package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jholhewres/goclaw/pkg/goclaw/copilot"
	"github.com/spf13/cobra"
)

// newSetupCmd creates the `copilot setup` command for interactive configuration.
func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Long: `Starts an interactive wizard to create your initial config.yaml.
Asks for assistant name, owner phone number, model, language, and other essentials.

Examples:
  copilot setup`,
		RunE: runSetup,
	}
}

// runSetup executes the interactive setup flow.
func runSetup(_ *cobra.Command, _ []string) error {
	return runInteractiveSetup()
}

// runInteractiveSetup guides the user through config creation step by step.
func runInteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)
	cfg := copilot.DefaultConfig()
	var apiKeyForEnv string
	var apiKeyForKeyring bool

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║         GoClaw Copilot — Setup Wizard        ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// ── Step 1: Assistant name ──
	fmt.Printf("1. Assistant name [%s]: ", cfg.Name)
	if name := readLine(reader); name != "" {
		cfg.Name = name
	}

	// ── Step 2: Trigger keyword ──
	fmt.Printf("2. Trigger keyword [%s]: ", cfg.Trigger)
	if trigger := readLine(reader); trigger != "" {
		cfg.Trigger = trigger
	}

	// ── Step 3: Owner phone number ──
	fmt.Println()
	fmt.Println("   The owner has full control over the bot.")
	fmt.Println("   Use your phone number with country code, no +, spaces or dashes.")
	fmt.Println("   Example: 5511999998888")
	fmt.Println()
	for {
		fmt.Print("3. Your phone number (owner): ")
		owner := readLine(reader)
		if owner == "" {
			fmt.Println("   [!] Phone number is required. The bot needs at least one owner.")
			continue
		}
		// Normalize: remove +, spaces, dashes.
		owner = normalizePhone(owner)
		if len(owner) < 10 {
			fmt.Println("   [!] Number seems too short. Include the country code (e.g. 5511999998888).")
			continue
		}
		cfg.Access.Owners = []string{owner}
		break
	}

	// ── Step 4: Access policy ──
	fmt.Println()
	fmt.Println("   Access policy for unknown contacts:")
	fmt.Println("   deny  — silently ignore (recommended)")
	fmt.Println("   allow — respond to everyone")
	fmt.Println("   ask   — send a one-time access request message")
	fmt.Println()
	fmt.Printf("4. Access policy [%s]: ", cfg.Access.DefaultPolicy)
	if policy := readLine(reader); policy != "" {
		switch strings.ToLower(policy) {
		case "deny", "allow", "ask":
			cfg.Access.DefaultPolicy = copilot.AccessPolicy(strings.ToLower(policy))
		default:
			fmt.Println("   [!] Invalid value, using 'deny'.")
		}
	}

	// ── Step 5: API provider ──
	fmt.Println()
	fmt.Println("   API endpoint (OpenAI-compatible):")
	fmt.Println()
	fmt.Printf("5. API base URL [%s]: ", cfg.API.BaseURL)
	if url := readLine(reader); url != "" {
		cfg.API.BaseURL = url
	}

	fmt.Println()
	hasKeyring := copilot.KeyringAvailable()
	if hasKeyring {
		fmt.Println("   API key — choose where to store it:")
		fmt.Println("     a) OS keyring (most secure — encrypted by the OS)")
		fmt.Println("     b) .env file (good — plaintext but gitignored)")
		fmt.Println("     c) Skip (set GOCLAW_API_KEY env var later)")
	} else {
		fmt.Println("   API key — choose how to provide it:")
		fmt.Println("     a) .env file (good — plaintext but gitignored)")
		fmt.Println("     b) Skip (set GOCLAW_API_KEY env var later)")
	}
	fmt.Println()
	fmt.Print("   API key (or press Enter to skip): ")
	if key := readLine(reader); key != "" {
		if hasKeyring {
			fmt.Print("   Store in OS keyring? (y/n) [y]: ")
			if ans := readLine(reader); ans == "" || strings.ToLower(ans) == "y" {
				if err := copilot.StoreKeyring("api_key", key); err != nil {
					fmt.Printf("   [!] Keyring failed: %v. Falling back to .env\n", err)
					apiKeyForEnv = key
				} else {
					fmt.Println("   API key stored in OS keyring (encrypted).")
					apiKeyForKeyring = true
				}
			} else {
				apiKeyForEnv = key
			}
		} else {
			apiKeyForEnv = key
		}
		cfg.API.APIKey = "${GOCLAW_API_KEY}"
	}

	// ── Step 6: Model (interactive numbered list) ──
	type modelOption struct {
		id   string
		name string
		desc string
	}

	models := []modelOption{
		// OpenAI
		{"gpt-5-mini", "GPT-5 Mini", "fast and cost-effective (default)"},
		{"gpt-5", "GPT-5", "latest OpenAI flagship"},
		{"gpt-4.5-preview", "GPT-4.5 Preview", "enhanced reasoning"},
		{"gpt-4o", "GPT-4o", "great all-around"},
		{"gpt-4o-mini", "GPT-4o Mini", "fast and cheap"},
		// Anthropic
		{"claude-opus-4.6", "Claude Opus 4.6", "most capable Anthropic"},
		{"claude-opus-4.5", "Claude Opus 4.5", "previous flagship"},
		{"claude-sonnet-4.5", "Claude Sonnet 4.5", "balanced performance"},
		// GLM (api.z.ai)
		{"glm-5", "GLM-5", "most capable GLM"},
		{"glm-4.7", "GLM-4.7", "balanced capability"},
		{"glm-4.7-flash", "GLM-4.7 Flash", "fast, low cost"},
		{"glm-4.7-flashx", "GLM-4.7 FlashX", "fast with extended context"},
	}

	// Find default index.
	defaultIdx := 0
	for i, m := range models {
		if m.id == cfg.Model {
			defaultIdx = i
			break
		}
	}

	fmt.Println()
	fmt.Println("6. Select LLM model:")
	fmt.Println()
	fmt.Println("   OpenAI:")
	for i := 0; i < 5; i++ {
		marker := "  "
		if i == defaultIdx {
			marker = " *"
		}
		fmt.Printf("   %s %2d) %-20s — %s\n", marker, i+1, models[i].id, models[i].desc)
	}
	fmt.Println()
	fmt.Println("   Anthropic:")
	for i := 5; i < 8; i++ {
		marker := "  "
		if i == defaultIdx {
			marker = " *"
		}
		fmt.Printf("   %s %2d) %-20s — %s\n", marker, i+1, models[i].id, models[i].desc)
	}
	fmt.Println()
	fmt.Println("   GLM (api.z.ai):")
	for i := 8; i < len(models); i++ {
		marker := "  "
		if i == defaultIdx {
			marker = " *"
		}
		fmt.Printf("   %s %2d) %-20s — %s\n", marker, i+1, models[i].id, models[i].desc)
	}
	fmt.Println()
	fmt.Printf("   Choose [1-%d] or type model name [%s]: ", len(models), cfg.Model)

	if input := readLine(reader); input != "" {
		// Check if it's a number.
		if idx, err := fmt.Sscanf(input, "%d", new(int)); idx == 1 && err == nil {
			var num int
			fmt.Sscanf(input, "%d", &num)
			if num >= 1 && num <= len(models) {
				cfg.Model = models[num-1].id
			} else {
				fmt.Printf("   [!] Invalid number, keeping '%s'.\n", cfg.Model)
			}
		} else {
			// Treat as raw model name.
			cfg.Model = input
		}
	}

	// Auto-adjust API base URL based on model choice.
	if strings.HasPrefix(cfg.Model, "glm-") && cfg.API.BaseURL == "https://api.openai.com/v1" {
		cfg.API.BaseURL = "https://api.z.ai/api/anthropic"
		fmt.Printf("   API URL auto-set to %s for GLM models.\n", cfg.API.BaseURL)
	} else if strings.HasPrefix(cfg.Model, "claude-") && cfg.API.BaseURL == "https://api.openai.com/v1" {
		cfg.API.BaseURL = "https://api.anthropic.com/v1"
		fmt.Printf("   API URL auto-set to %s for Anthropic models.\n", cfg.API.BaseURL)
	}

	// ── Step 7: Language ──
	fmt.Printf("7. Response language [%s]: ", cfg.Language)
	if lang := readLine(reader); lang != "" {
		cfg.Language = lang
	}

	// ── Step 8: Timezone ──
	fmt.Printf("8. Timezone [%s]: ", cfg.Timezone)
	if tz := readLine(reader); tz != "" {
		cfg.Timezone = tz
	}

	// ── Step 9: System instructions ──
	fmt.Println()
	fmt.Println("   Base system instructions (system prompt).")
	fmt.Println("   Press Enter to keep the default.")
	fmt.Println()
	fmt.Printf("9. Instructions [default]: ")
	if instr := readLine(reader); instr != "" {
		cfg.Instructions = instr
	}

	// ── Step 10: WhatsApp settings ──
	fmt.Println()
	fmt.Println("   WhatsApp settings:")
	fmt.Println()

	fmt.Printf("   Respond in groups? (y/n) [y]: ")
	if g := readLine(reader); strings.ToLower(g) == "n" {
		cfg.Channels.WhatsApp.RespondToGroups = false
	}

	fmt.Printf("   Respond in DMs? (y/n) [y]: ")
	if d := readLine(reader); strings.ToLower(d) == "n" {
		cfg.Channels.WhatsApp.RespondToDMs = false
	}

	// ── Summary ──
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println("  Configuration summary:")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Printf("  Name:      %s\n", cfg.Name)
	fmt.Printf("  Trigger:   %s\n", cfg.Trigger)
	fmt.Printf("  Owner:     %s\n", cfg.Access.Owners[0])
	fmt.Printf("  Policy:    %s\n", cfg.Access.DefaultPolicy)
	fmt.Printf("  API URL:   %s\n", cfg.API.BaseURL)
	if apiKeyForKeyring {
		fmt.Printf("  API key:   **** (→ OS keyring)\n")
	} else if apiKeyForEnv != "" {
		fmt.Printf("  API key:   ****%s (→ .env)\n", apiKeyForEnv[max(0, len(apiKeyForEnv)-4):])
	} else {
		fmt.Printf("  API key:   (set GOCLAW_API_KEY later)\n")
	}
	fmt.Printf("  Model:     %s\n", cfg.Model)
	fmt.Printf("  Language:  %s\n", cfg.Language)
	fmt.Printf("  Timezone:  %s\n", cfg.Timezone)
	fmt.Printf("  Groups:    %v\n", cfg.Channels.WhatsApp.RespondToGroups)
	fmt.Printf("  DMs:       %v\n", cfg.Channels.WhatsApp.RespondToDMs)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// ── Confirm and save ──
	target := "config.yaml"
	fmt.Printf("Save to %s? (y/n) [y]: ", target)
	if confirm := readLine(reader); strings.ToLower(confirm) == "n" {
		fmt.Println("Setup cancelled.")
		return nil
	}

	// Check if already exists.
	if _, err := os.Stat(target); err == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/n) [n]: ", target)
		if overwrite := readLine(reader); strings.ToLower(overwrite) != "y" {
			fmt.Println("Setup cancelled. Existing file kept.")
			return nil
		}
	}

	if err := copilot.SaveConfigToFile(cfg, target); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Save API key to .env file (never in config.yaml).
	if apiKeyForEnv != "" {
		envContent := fmt.Sprintf("# GoClaw secrets — DO NOT commit this file.\n# It is already in .gitignore.\nGOCLAW_API_KEY=%s\n", apiKeyForEnv)
		if err := os.WriteFile(".env", []byte(envContent), 0o600); err != nil {
			fmt.Printf("   [!] Failed to write .env: %v\n", err)
			fmt.Printf("   Set manually: export GOCLAW_API_KEY=%s\n", apiKeyForEnv)
		} else {
			fmt.Println(".env created with your API key (permissions: 600).")
		}
	}

	fmt.Printf("\nconfig.yaml created successfully!\n\n")

	fmt.Println("Security notes:")
	fmt.Println("  - config.yaml and .env are in .gitignore (never committed)")
	if apiKeyForKeyring {
		fmt.Println("  - API key is encrypted in the OS keyring (most secure)")
	} else if apiKeyForEnv != "" {
		fmt.Println("  - API key is stored in .env, not in config.yaml")
		fmt.Println("  - For maximum security, run: copilot config set-key")
	}
	fmt.Println("  - config.yaml permissions: 600 (owner only)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review config.yaml and adjust as needed")
	fmt.Println("  2. Run: copilot serve")
	fmt.Println("  3. Scan the QR code with your WhatsApp")
	fmt.Println()

	return nil
}

// readLine reads a single line from the reader, trimming whitespace.
func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// normalizePhone removes common phone number formatting characters.
func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, "+", "")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}
