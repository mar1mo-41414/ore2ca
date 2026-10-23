package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/mar1mo-41414/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	var (
		purge bool
		yes   bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "CAの信頼を解除する（--purge で完全削除）",
		Long: `OS・ブラウザの信頼ストアからCAを削除します。

--purge を指定すると、信頼解除に加えて ~/.ore2ca ディレクトリ（CA証明書・秘密鍵・
発行済み証明書・設定ファイルすべて）を削除します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません")
			}

			fmt.Println("CAをシステムの信頼ストアから削除中...")
			result, err := trust.Uninstall(s, trust.Options{})
			if err != nil {
				return err
			}
			printUntrustResult(result)

			if !purge {
				return untrustExitError(result)
			}

			baseDir := s.BaseDir()

			// OS信頼ストア削除失敗時は再確認
			if !result.OS && !yes {
				fmt.Printf("\n✗ OS信頼ストアの削除に失敗しました。\n")
				fmt.Printf("  それでも %s を削除しますか？ [y/N]: ", baseDir)
				reader := bufio.NewReader(os.Stdin)
				line, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(line), "y") {
					fmt.Println("キャンセルしました")
					return fmt.Errorf("OS信頼ストアからの削除に失敗しました")
				}
			}

			if !yes {
				fmt.Printf("\n⚠ %s を完全に削除します。\n", baseDir)
				fmt.Println("  CA証明書・秘密鍵・発行済み証明書がすべて失われます。この操作は取り消せません。")
				fmt.Printf("  続行しますか？ [y/N]: ")
				reader := bufio.NewReader(os.Stdin)
				line, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(line), "y") {
					fmt.Println("キャンセルしました")
					return nil
				}
			}

			if err := os.RemoveAll(baseDir); err != nil {
				return fmt.Errorf("✗ %s の削除に失敗しました: %w", baseDir, err)
			}
			fmt.Printf("✓ %s を削除しました\n", baseDir)
			return nil
		},
	}

	cmd.Flags().BoolVar(&purge, "purge", false, "~/.ore2ca ディレクトリを完全に削除する")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "確認をスキップ")
	return cmd
}
