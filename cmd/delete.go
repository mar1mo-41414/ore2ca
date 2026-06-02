package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ore2ca/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "証明書を削除する",
		Long:  `指定したIDの証明書とその関連ファイルをすべて削除します。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			s, err := store.New()
			if err != nil {
				return err
			}

			meta, err := s.FindByID(id)
			if err != nil {
				return err
			}

			if !yes {
				fmt.Printf("証明書を削除しますか？\n  ID: %s\n  ドメイン: %s\n[y/N]: ", meta.ID, meta.Domain)
				reader := bufio.NewReader(os.Stdin)
				line, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(line), "y") {
					fmt.Println("キャンセルしました")
					return nil
				}
			}

			if err := s.DeleteCert(meta.Domain); err != nil {
				return fmt.Errorf("削除失敗: %w", err)
			}

			fmt.Printf("✓ 証明書を削除しました: ID=%s, ドメイン=%s\n", meta.ID, meta.Domain)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "確認をスキップ")
	return cmd
}
