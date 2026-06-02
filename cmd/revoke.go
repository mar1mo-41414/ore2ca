package cmd

import (
	"fmt"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/ca"
	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <id>",
		Short: "証明書を失効させる",
		Long:  `指定したIDの証明書を失効させます。証明書ファイルは削除されません。`,
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

			if meta.Revoked {
				return fmt.Errorf("証明書 %s はすでに失効済みです", id)
			}

			meta.Revoked = true
			meta.RevokedAt = time.Now()

			if err := s.SaveMeta(meta); err != nil {
				return err
			}

			fmt.Printf("✓ 証明書を失効させました: ID=%s, ドメイン=%s\n", meta.ID, meta.Domain)

			// CRL を自動再生成
			if err := ca.BuildCRL(s); err != nil {
				fmt.Printf("⚠ CRL の更新に失敗しました: %v\n", err)
			} else {
				fmt.Printf("  CRL:        %s\n", s.CRLPath())
			}

			return nil
		},
	}
	return cmd
}
