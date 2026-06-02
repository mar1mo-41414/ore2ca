package cmd

import (
	"fmt"

	"github.com/ore2ca/ore2ca/internal/ca"
	"github.com/ore2ca/ore2ca/internal/config"
	"github.com/ore2ca/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		commonName   string
		organization string
		country      string
		validYears   int
		force        bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "ローカルCAを作成する",
		Long:  `新しいローカルルートCAを作成します。~/.ore2ca/ に保存されます。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if commonName != "" {
				cfg.CA.CommonName = commonName
			}
			if organization != "" {
				cfg.CA.Organization = organization
			}
			if country != "" {
				cfg.CA.Country = country
			}
			if validYears > 0 {
				cfg.CA.ValidYears = validYears
			}

			s, err := store.New()
			if err != nil {
				return err
			}

			if s.CAExists() && !force {
				return fmt.Errorf("CAはすでに存在します。上書きするには --force を使用してください")
			}

			fmt.Printf("ローカルCA を作成中: %s\n", cfg.CA.CommonName)
			if err := ca.Init(cfg, s); err != nil {
				return fmt.Errorf("CA作成失敗: %w", err)
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Println("✓ CA作成完了")
			fmt.Printf("  証明書: %s\n", s.CACertPath())
			fmt.Printf("  秘密鍵: %s\n", s.CAKeyPath())
			fmt.Println()
			fmt.Println("次のステップ: ore2ca trust  # OSに信頼登録")
			return nil
		},
	}

	cmd.Flags().StringVar(&commonName, "cn", "", "CA の Common Name")
	cmd.Flags().StringVar(&organization, "org", "", "CA の組織名")
	cmd.Flags().StringVar(&country, "country", "", "CA の国コード (例: JP)")
	cmd.Flags().IntVar(&validYears, "years", 0, "CA 証明書の有効期間（年）")
	cmd.Flags().BoolVar(&force, "force", false, "既存のCAを上書き")

	return cmd
}
