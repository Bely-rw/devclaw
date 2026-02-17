package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newShellHookCmd creates the `devclaw shell-hook` command that generates
// shell integration scripts to auto-capture errors.
func newShellHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell-hook [bash|zsh|fish]",
		Short: "Generate shell hook for automatic error capture",
		Long: `Generate a shell integration script that captures command errors
and offers to analyze them with DevClaw.

To install:
  eval "$(devclaw shell-hook bash)"    # add to ~/.bashrc
  eval "$(devclaw shell-hook zsh)"     # add to ~/.zshrc
  devclaw shell-hook fish | source     # add to ~/.config/fish/config.fish`,
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				fmt.Print(bashHook)
			case "zsh":
				fmt.Print(zshHook)
			case "fish":
				fmt.Print(fishHook)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", args[0])
			}
			return nil
		},
	}
	return cmd
}

const bashHook = `# DevClaw shell hook — auto-capture errors
# Add to ~/.bashrc: eval "$(devclaw shell-hook bash)"
__devclaw_prompt_command() {
  local exit_code=$?
  if [ $exit_code -ne 0 ] && [ $exit_code -ne 130 ]; then
    local last_cmd=$(HISTTIMEFORMAT='' history 1 | sed 's/^[ ]*[0-9]*[ ]*//')
    echo -e "\033[33m[devclaw]\033[0m Command failed (exit $exit_code): $last_cmd"
    echo -e "\033[33m[devclaw]\033[0m Run: devclaw fix"
    export DEVCLAW_LAST_ERROR="$last_cmd (exit $exit_code)"
  fi
}
PROMPT_COMMAND="__devclaw_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`

const zshHook = `# DevClaw shell hook — auto-capture errors
# Add to ~/.zshrc: eval "$(devclaw shell-hook zsh)"
__devclaw_precmd() {
  local exit_code=$?
  if [[ $exit_code -ne 0 ]] && [[ $exit_code -ne 130 ]]; then
    local last_cmd=$(fc -ln -1)
    echo -e "\033[33m[devclaw]\033[0m Command failed (exit $exit_code): $last_cmd"
    echo -e "\033[33m[devclaw]\033[0m Run: devclaw fix"
    export DEVCLAW_LAST_ERROR="$last_cmd (exit $exit_code)"
  fi
}
precmd_functions+=(__devclaw_precmd)
`

const fishHook = `# DevClaw shell hook — auto-capture errors
# Add to config.fish: devclaw shell-hook fish | source
function __devclaw_postexec --on-event fish_postexec
  set -l exit_code $status
  if test $exit_code -ne 0; and test $exit_code -ne 130
    set -l last_cmd $argv[1]
    echo -e "\033[33m[devclaw]\033[0m Command failed (exit $exit_code): $last_cmd"
    echo -e "\033[33m[devclaw]\033[0m Run: devclaw fix"
    set -gx DEVCLAW_LAST_ERROR "$last_cmd (exit $exit_code)"
  end
end
`
