package cmd

import (
	"fmt"

	"github.com/ore2ca/ore2ca/internal/ca"
	"github.com/ore2ca/ore2ca/internal/config"
	"github.com/ore2ca/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newIssueCmd() *cobra.Command {
	var validDays int

	cmd := &cobra.Command{
		Use:   "issue <domain>",
		Short: "サーバ証明書を発行する",
		Long: `指定したドメインのサーバ証明書を発行します。

例:
  ore2ca issue localhost
  ore2ca issue jellyfin.home.arpa
  ore2ca issue '*.home.arpa'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]

			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません。先に ore2ca init を実行してください")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if validDays > 0 {
				cfg.Certs.ValidDays = validDays
			}

			fmt.Printf("証明書を発行中: %s\n", domain)
			meta, err := ca.Issue(domain, cfg, s)
			if err != nil {
				return fmt.Errorf("証明書発行失敗: %w", err)
			}

			fmt.Println("✓ 証明書発行完了")
			fmt.Printf("  ID:       %s\n", meta.ID)
			fmt.Printf("  ドメイン: %s\n", meta.Domain)
			fmt.Printf("  有効期限: %s\n", meta.ExpiresAt.Format("2006-01-02"))
			fmt.Printf("  証明書:   %s\n", s.CertPath(domain))
			fmt.Printf("  秘密鍵:   %s\n", s.KeyPath(domain))
			fmt.Printf("  フルチェーン: %s\n", s.FullchainPath(domain))
			return nil
		},
	}

	cmd.Flags().IntVar(&validDays, "days", 0, "証明書の有効期間（日）")
	return cmd
}
