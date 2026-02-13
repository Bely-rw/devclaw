package commands

import (
	"fmt"
	"os"

	"github.com/jholhewres/goclaw/pkg/goclaw/copilot"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConfigCmd creates the `copilot config` command.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage assistant configuration",
		Long: `Manage GoClaw Copilot configuration.

Examples:
  copilot config init
  copilot config show
  copilot config validate`,
	}

	cmd.AddCommand(
		newConfigInitCmd(),
		newConfigShowCmd(),
		newConfigValidateCmd(),
	)

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default config.yaml",
		RunE: func(_ *cobra.Command, _ []string) error {
			target := "config.yaml"

			// Check if already exists.
			if _, err := os.Stat(target); err == nil {
				return fmt.Errorf("config.yaml already exists. Remove it first or edit it directly")
			}

			// Write default config.
			cfg := copilot.DefaultConfig()
			if err := copilot.SaveConfigToFile(cfg, target); err != nil {
				return err
			}

			fmt.Printf("Created %s with default configuration.\n", target)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Edit config.yaml and set your phone number in access.owners")
			fmt.Println("  2. Run: copilot serve")
			fmt.Println("  3. Scan the QR code with WhatsApp")
			return nil
		},
	}
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, path, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("# Loaded from: %s\n\n", path)

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, path, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Config: %s\n", path)
			fmt.Printf("  Name:      %s\n", cfg.Name)
			fmt.Printf("  Model:     %s\n", cfg.Model)
			fmt.Printf("  Trigger:   %s\n", cfg.Trigger)
			fmt.Printf("  Language:  %s\n", cfg.Language)
			fmt.Printf("  Policy:    %s\n", cfg.Access.DefaultPolicy)
			fmt.Printf("  Owners:    %d\n", len(cfg.Access.Owners))
			fmt.Printf("  Admins:    %d\n", len(cfg.Access.Admins))
			fmt.Printf("  Users:     %d\n", len(cfg.Access.AllowedUsers))

			wsCount := len(cfg.Workspaces.Workspaces)
			fmt.Printf("  Workspaces: %d\n", wsCount)
			for _, ws := range cfg.Workspaces.Workspaces {
				fmt.Printf("    - %s (%s): %d members, %d groups\n",
					ws.ID, ws.Name, len(ws.Members), len(ws.Groups))
			}

			fmt.Println("\nConfiguration is valid.")
			return nil
		},
	}
}

// loadConfig loads the config from the --config flag or auto-discovers it.
func loadConfig(cmd *cobra.Command) (*copilot.Config, string, error) {
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	if configPath == "" {
		configPath = copilot.FindConfigFile()
	}

	if configPath == "" {
		return nil, "", fmt.Errorf("no config file found.\nRun 'copilot config init' to create one, or use --config <path>")
	}

	cfg, err := copilot.LoadConfigFromFile(configPath)
	if err != nil {
		return nil, configPath, fmt.Errorf("loading config from %s: %w", configPath, err)
	}

	return cfg, configPath, nil
}
