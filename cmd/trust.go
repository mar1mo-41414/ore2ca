package cmd

import (
	"fmt"

	"github.com/ore2ca/ore2ca/internal/store"
	"github.com/ore2ca/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newTrustCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust",
		Short: "CAをOSの信頼ストアに登録する",
		Long:  `作成済みのローカルCAをOSのルート証明書ストアに登録します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません。先に ore2ca init を実行してください")
			}

			fmt.Println("CAをシステムの信頼ストアに登録中...")
			result, err := trust.Install(s)
			if err != nil {
				return err
			}

			if result.OS {
				fmt.Println("✓ OS信頼ストア: 登録完了")
			} else {
				fmt.Printf("✗ OS信頼ストア: 失敗 - %v\n", result.OSErr)
			}

			if result.Firefox {
				fmt.Println("✓ Firefox NSS: 登録完了")
			} else if result.FFErr != nil {
				fmt.Printf("  Firefox NSS: スキップ - %v\n", result.FFErr)
			}

			if !result.OS {
				return fmt.Errorf("OS信頼登録に失敗しました")
			}
			return nil
		},
	}
	return cmd
}
