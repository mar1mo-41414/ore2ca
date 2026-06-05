package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "シェル補完スクリプトを生成する",
		Long: `指定したシェル用の補完スクリプトを生成します。

Bash:

  $ source <(ore2ca completion bash)

  # セッションをまたいで有効にするには、一度だけ実行:
  # Linux:
  $ ore2ca completion bash > /etc/bash_completion.d/ore2ca
  # macOS:
  $ ore2ca completion bash > $(brew --prefix)/etc/bash_completion.d/ore2ca

Zsh:

  # Zsh の補完が有効でない場合は先に実行:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  $ source <(ore2ca completion zsh)

  # セッションをまたいで有効にするには、一度だけ実行:
  $ ore2ca completion zsh > "${fpath[1]}/_ore2ca"

Fish:

  $ ore2ca completion fish | source

  # セッションをまたいで有効にするには、一度だけ実行:
  $ ore2ca completion fish > ~/.config/fish/completions/ore2ca.fish

PowerShell:

  PS> ore2ca completion powershell | Out-String | Invoke-Expression

  # セッションをまたいで有効にするには、$PROFILE に追記:
  PS> ore2ca completion powershell >> $PROFILE
`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
}
