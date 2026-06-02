package cmd

import (
	"fmt"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/mar1mo-41414/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newUntrustCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrust",
		Short: "CAをOSの信頼ストアから削除する",
		Long:  `ore2ca trust で登録したCA証明書をOSの信頼ストアおよびFirefox NSSから削除します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません")
			}

			fmt.Println("CAをシステムの信頼ストアから削除中...")
			result, err := trust.Uninstall(s)
			if err != nil {
				return err
			}

			if result.OS {
				fmt.Println("✓ OS信頼ストア: 削除完了")
			} else {
				fmt.Printf("✗ OS信頼ストア: 失敗 - %v\n", result.OSErr)
			}

			if result.Firefox {
				fmt.Println("✓ Firefox NSS: 削除完了")
			} else if result.FFErr != nil {
				fmt.Printf("✗ Firefox NSS: 削除失敗\n\n%v\n", result.FFErr)
			}

			if !result.OS {
				return fmt.Errorf("\nOS信頼ストアからの削除に失敗しました")
			}

			fmt.Println("\nブラウザを再起動すると変更が反映されます。")
			return nil
		},
	}
	return cmd
}
