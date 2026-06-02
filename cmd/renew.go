package cmd

import (
	"fmt"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/ca"
	"github.com/mar1mo-41414/ore2ca/internal/config"
	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newRenewCmd() *cobra.Command {
	var validDays int

	cmd := &cobra.Command{
		Use:   "renew <id>",
		Short: "証明書を更新する",
		Long: `指定したIDの証明書を同じドメイン・SANで再発行します。

既存の証明書ファイルは上書きされます。
有効期間を変えたい場合は --days を指定してください。`,
		Args: cobra.ExactArgs(1),
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

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if validDays > 0 {
				cfg.Certs.ValidDays = validDays
			}

			if len(meta.SANs) > 0 {
				fmt.Printf("証明書を更新中: %s (追加SAN: %s)\n", meta.Domain, strings.Join(meta.SANs, ", "))
			} else {
				fmt.Printf("証明書を更新中: %s\n", meta.Domain)
			}

			newMeta, err := ca.Issue(meta.Domain, cfg, s, meta.SANs...)
			if err != nil {
				return fmt.Errorf("証明書更新失敗: %w", err)
			}

			fmt.Println("✓ 証明書更新完了")
			fmt.Printf("  ID:         %s → %s\n", meta.ID, newMeta.ID)
			fmt.Printf("  ドメイン:   %s\n", newMeta.Domain)
			fmt.Printf("  旧有効期限: %s\n", meta.ExpiresAt.Format("2006-01-02"))
			fmt.Printf("  新有効期限: %s\n", newMeta.ExpiresAt.Format("2006-01-02"))
			fmt.Printf("  証明書:     %s\n", s.CertPath(newMeta.Domain))
			return nil
		},
	}

	cmd.Flags().IntVar(&validDays, "days", 0, "証明書の有効期間（日）省略時は設定値を使用")
	return cmd
}
