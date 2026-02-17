package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// newCompletionCmd creates the `devclaw completion` command that generates
// shell completion scripts for bash, zsh, fish, and powershell.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell auto-completion scripts for devclaw.

To load completions:

Bash:
  $ source <(devclaw completion bash)
  # To load completions for each session, add to ~/.bashrc:
  echo 'source <(devclaw completion bash)' >> ~/.bashrc

Zsh:
  $ source <(devclaw completion zsh)
  # To load completions for each session, add to ~/.zshrc:
  echo 'source <(devclaw completion zsh)' >> ~/.zshrc

Fish:
  $ devclaw completion fish | source
  # To load completions for each session:
  devclaw completion fish > ~/.config/fish/completions/devclaw.fish

PowerShell:
  PS> devclaw completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, add to your profile:
  devclaw completion powershell | Out-String | Invoke-Expression`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
