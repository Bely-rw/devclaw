package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newExplainCmd creates the `devclaw explain` command that explains
// a file, directory, or codebase structure.
func newExplainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain [path]",
		Short: "Explain code, files, or directories",
		Long: `Explain the purpose and structure of a file, directory, or codebase.

Examples:
  devclaw explain .                    # explain current project
  devclaw explain ./src/auth/          # explain auth module
  devclaw explain main.go              # explain a file`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := resolveConfig(cmd)
			if err != nil {
				return err
			}

			assistant, cleanup, err := quickAssistant(cfg, cmd)
			if err != nil {
				return err
			}
			defer cleanup()

			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("path not found: %s", target)
			}

			var prompt string
			if info.IsDir() {
				files := listDirTree(target, 3)
				prompt = fmt.Sprintf("Explain the structure and purpose of this directory:\n\nPath: %s\n\n```\n%s\n```", target, files)
			} else {
				content, err := os.ReadFile(target)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
				prompt = fmt.Sprintf("Explain this code — what it does, its purpose, and key patterns:\n\nFile: %s\n```\n%s\n```", target, string(content))
			}

			response := executeChat(assistant, prompt)
			fmt.Println(response)
			return nil
		},
	}
	return cmd
}

// listDirTree returns a simple tree representation of a directory.
func listDirTree(root string, maxDepth int) string {
	var sb strings.Builder
	walkDir(root, "", 0, maxDepth, &sb)
	return sb.String()
}

func walkDir(path, prefix string, depth, maxDepth int, sb *strings.Builder) {
	if depth >= maxDepth {
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
			continue
		}
		sb.WriteString(prefix + name)
		if e.IsDir() {
			sb.WriteString("/\n")
			walkDir(filepath.Join(path, name), prefix+"  ", depth+1, maxDepth, sb)
		} else {
			sb.WriteString("\n")
		}
	}
}
