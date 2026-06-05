package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ore2ca",
	Short: "俺俺CA - ローカル認証局 & HTTPS開発環境ツール",
	Long: `ore2ca はローカル開発用の認証局（CA）を管理し、
HTTPS証明書の発行・信頼登録・管理をワンストップで行うツールです。`,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// cobra のデフォルト completion コマンドを無効化し、
	// インストール手順つきのカスタム版に差し替える
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(
		newInitCmd(),
		newImportCmd(),
		newTrustCmd(),
		newUntrustCmd(),
		newIssueCmd(),
		newRenewCmd(),
		newShowCmd(),
		newExportCmd(),
		newListCmd(),
		newRevokeCmd(),
		newDeleteCmd(),
		newDockerCmd(),
		newWebCmd(),
		newCompletionCmd(rootCmd),
	)
}
