package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newHowCmd creates the `devclaw how` command that generates shell commands
// for a given task without executing them.
func newHowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "how <task description>",
		Short: "Generate shell commands for a task (without executing)",
		Long: `Describe what you want to do and get the shell commands, without
executing them. Useful for learning and confirming before running.

Examples:
  devclaw how "compress all log files in /var/log"
  devclaw how "find large files over 100MB"
  devclaw how "set up a PostgreSQL database"`,
		Args: cobra.MinimumNArgs(1),
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

			task := strings.Join(args, " ")
			prompt := fmt.Sprintf(`Generate the shell command(s) to accomplish this task. Return ONLY the commands, one per line, with brief comments if needed. Do NOT execute anything. Do NOT wrap in markdown code blocks.

Task: %s`, task)

			response := executeChat(assistant, prompt)
			fmt.Println(response)
			return nil
		},
	}
	return cmd
}
