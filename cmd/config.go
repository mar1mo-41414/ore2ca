package cmd

import (
	"fmt"

	"github.com/mar1mo-41414/ore2ca/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "設定の表示・変更",
		Long:  `ore2ca の設定（~/.ore2ca/config.yaml）を表示・変更します。`,
	}

	cmd.AddCommand(newConfigShowCmd(), newConfigSetCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "現在の設定を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			fmt.Println("=== ore2ca 設定 ===")
			fmt.Println()
			fmt.Println("[CA]")
			fmt.Printf("  Common Name  : %s\n", cfg.CA.CommonName)
			fmt.Printf("  Organization : %s\n", cfg.CA.Organization)
			fmt.Printf("  Country      : %s\n", cfg.CA.Country)
			fmt.Printf("  Valid Years  : %d 年\n", cfg.CA.ValidYears)
			fmt.Println()
			fmt.Println("[Certs]")
			fmt.Printf("  Default Valid Days : %d 日\n", cfg.Certs.ValidDays)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "設定値を変更する",
		Long: `設定値を変更して ~/.ore2ca/config.yaml に保存します。

使用例:
  ore2ca config set cert-days 365   サーバ証明書のデフォルト有効期間を365日に変更`,
	}

	cmd.AddCommand(newConfigSetCertDaysCmd())
	return cmd
}

func newConfigSetCertDaysCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cert-days <n>",
		Short: "サーバ証明書のデフォルト有効期間（日数）を設定する",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var days int
			if _, err := fmt.Sscanf(args[0], "%d", &days); err != nil || days <= 0 {
				return fmt.Errorf("無効な値です: %q（1以上の整数を指定してください）", args[0])
			}
			if days > 36500 {
				return fmt.Errorf("有効期間が長すぎます: %d 日（最大 36500 日 = 100年）", days)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.Certs.ValidDays = days
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Printf("✓ サーバ証明書のデフォルト有効期間を %d 日に設定しました\n", days)
			fmt.Println("  （既存の証明書には影響しません。次回 ore2ca issue から適用されます）")
			return nil
		},
	}
}
