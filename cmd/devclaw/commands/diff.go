package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// newDiffCmd creates the `devclaw diff` command that reviews git changes.
func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Review git diff with AI analysis",
		Long: `Review the current git diff and get AI analysis: potential issues,
suggestions, and a summary of changes.

Examples:
  devclaw diff             # review unstaged changes
  devclaw diff --staged    # review staged changes`,
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

			staged, _ := cmd.Flags().GetBool("staged")

			gitArgs := []string{"diff"}
			if staged {
				gitArgs = append(gitArgs, "--cached")
			}

			out, err := exec.Command("git", gitArgs...).CombinedOutput()
			if err != nil {
				return fmt.Errorf("git diff failed: %s", strings.TrimSpace(string(out)))
			}

			diffContent := strings.TrimSpace(string(out))
			if diffContent == "" {
				fmt.Println("No changes to review.")
				return nil
			}

			prompt := fmt.Sprintf("Review this git diff. Identify potential issues, suggest improvements, and provide a brief summary:\n\n```diff\n%s\n```", diffContent)

			response := executeChat(assistant, prompt)
			fmt.Println(response)
			return nil
		},
	}

	cmd.Flags().Bool("staged", false, "review staged changes")
	return cmd
}
