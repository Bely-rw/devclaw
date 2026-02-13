package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jholhewres/goclaw/pkg/goclaw/channels/whatsapp"
	"github.com/jholhewres/goclaw/pkg/goclaw/copilot"
	"github.com/jholhewres/goclaw/pkg/goclaw/plugins"
	"github.com/spf13/cobra"
)

// newServeCmd creates the `copilot serve` command that starts the daemon.
func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the daemon with messaging channels",
		Long: `Start GoClaw Copilot as a daemon service, connecting to enabled
channels (WhatsApp, Discord, Telegram) and processing messages.

Examples:
  copilot serve
  copilot serve --channel whatsapp
  copilot serve --config ./config.yaml`,
		RunE: runServe,
	}

	cmd.Flags().StringSlice("channel", nil, "channels to enable (whatsapp, discord, telegram)")
	return cmd
}

func runServe(cmd *cobra.Command, _ []string) error {
	// ── Load config ──
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return err
	}

	// ── Configure logger ──
	verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")
	logLevel := slog.LevelInfo
	if verbose || cfg.Logging.Level == "debug" {
		logLevel = slog.LevelDebug
	}

	var handler slog.Handler
	if cfg.Logging.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}
	logger := slog.New(handler)

	// ── Create assistant ──
	assistant := copilot.New(cfg, logger)

	// ── Create context ──
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Register channels ──
	channelFilter, _ := cmd.Flags().GetStringSlice("channel")

	// WhatsApp (core channel).
	if shouldEnable("whatsapp", channelFilter, true) {
		wa := whatsapp.New(cfg.Channels.WhatsApp, logger)
		if err := assistant.ChannelManager().Register(wa); err != nil {
			logger.Error("failed to register WhatsApp", "error", err)
		} else {
			logger.Info("WhatsApp channel registered")
		}
	}

	// Load plugins (Discord, Telegram, etc.).
	pluginLoader := plugins.NewLoader(cfg.Plugins, logger)
	if err := pluginLoader.LoadAll(ctx); err != nil {
		logger.Error("failed to load plugins", "error", err)
	} else if pluginLoader.Count() > 0 {
		if err := pluginLoader.RegisterChannels(assistant.ChannelManager()); err != nil {
			logger.Error("failed to register plugin channels", "error", err)
		}
	}

	// ── Start ──
	if err := assistant.Start(ctx); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	// ── Wait for shutdown ──
	logger.Info("GoClaw Copilot running. Press Ctrl+C to stop.",
		"name", cfg.Name,
		"trigger", cfg.Trigger,
		"policy", cfg.Access.DefaultPolicy,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutdown signal received, stopping...")
	pluginLoader.Shutdown()
	assistant.Stop()

	return nil
}

// resolveConfig loads config from file or uses defaults.
func resolveConfig(cmd *cobra.Command) (*copilot.Config, error) {
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	// Try explicit path first.
	if configPath != "" {
		cfg, err := copilot.LoadConfigFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
		return cfg, nil
	}

	// Auto-discover config file.
	if found := copilot.FindConfigFile(); found != "" {
		cfg, err := copilot.LoadConfigFromFile(found)
		if err != nil {
			return nil, fmt.Errorf("loading config from %s: %w", found, err)
		}
		slog.Info("config loaded", "path", found)
		return cfg, nil
	}

	// No config file — use defaults.
	slog.Info("no config file found, using defaults")
	return copilot.DefaultConfig(), nil
}

// shouldEnable checks if a channel should be enabled.
func shouldEnable(name string, filter []string, defaultEnabled bool) bool {
	if len(filter) == 0 {
		return defaultEnabled
	}
	for _, f := range filter {
		if f == name {
			return true
		}
	}
	return false
}
