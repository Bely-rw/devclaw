package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// newCommitCmd creates the `devclaw commit` command that generates
// a commit message from staged changes and commits.
func newCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate commit message and commit staged changes",
		Long: `Analyze staged git changes and generate a conventional commit message,
then commit with that message.

Examples:
  devclaw commit           # generate message + commit
  devclaw commit --dry-run # generate message only, don't commit`,
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

			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Get staged diff
			out, err := exec.Command("git", "diff", "--cached", "--stat").CombinedOutput()
			if err != nil || strings.TrimSpace(string(out)) == "" {
				return fmt.Errorf("no staged changes. Stage files with: git add <files>")
			}
			stat := strings.TrimSpace(string(out))

			diffOut, _ := exec.Command("git", "diff", "--cached").CombinedOutput()
			diffContent := strings.TrimSpace(string(diffOut))

			// Truncate very long diffs
			const maxDiffLen = 6000
			if len(diffContent) > maxDiffLen {
				diffContent = diffContent[:maxDiffLen] + "\n... (truncated)"
			}

			prompt := fmt.Sprintf(`Generate a concise conventional commit message for these staged changes.
Use format: type(scope): description

Types: feat, fix, refactor, docs, style, test, chore, perf, ci, build
Scope is optional. Description should be imperative mood, lowercase, no period.

Return ONLY the commit message, nothing else.

Stats:
%s

Diff:
%s`, stat, diffContent)

			message := strings.TrimSpace(executeChat(assistant, prompt))

			// Clean up: remove backticks or quotes that LLM might add
			message = strings.Trim(message, "`\"'")
			message = strings.TrimSpace(message)

			fmt.Printf("Commit message: %s\n", message)

			if dryRun {
				return nil
			}

			commitOut, err := exec.Command("git", "commit", "-m", message).CombinedOutput()
			if err != nil {
				return fmt.Errorf("git commit failed: %s", strings.TrimSpace(string(commitOut)))
			}
			fmt.Println(strings.TrimSpace(string(commitOut)))
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "generate message only, don't commit")
	return cmd
}
