package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/mar1mo-41414/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		certPath string
		force    bool
		doTrust  bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "別PCで作成したCA証明書をインポートする",
		Long: `別のマシンで作成したCA証明書をインポートします。
インポート後、ore2ca trust で信頼登録できます。

例:
  ore2ca import --cert /path/to/root.crt
  ore2ca import --cert /path/to/root.crt --trust`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if certPath == "" {
				return fmt.Errorf("--cert にCA証明書のパスを指定してください")
			}

			data, err := os.ReadFile(certPath)
			if err != nil {
				return fmt.Errorf("証明書ファイルを読み込めません: %w", err)
			}

			block, _ := pem.Decode(data)
			if block == nil {
				return fmt.Errorf("PEM形式の証明書ではありません: %s", certPath)
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("証明書の解析に失敗しました: %w", err)
			}
			if !cert.IsCA {
				return fmt.Errorf("この証明書はCA証明書ではありません (IsCA=false)")
			}

			s, err := store.New()
			if err != nil {
				return err
			}
			if s.CAExists() && !force {
				return fmt.Errorf("CAはすでに存在します。上書きするには --force を使用してください")
			}

			if err := os.WriteFile(s.CACertPath(), data, 0644); err != nil {
				return fmt.Errorf("証明書の保存に失敗しました: %w", err)
			}

			fmt.Println("✓ CA証明書をインポートしました")
			fmt.Printf("  発行者: %s\n", cert.Subject.CommonName)
			fmt.Printf("  有効期限: %s\n", cert.NotAfter.Format("2006-01-02"))
			fmt.Printf("  保存先: %s\n", s.CACertPath())

			if doTrust {
				fmt.Println()
				fmt.Println("CAをシステムの信頼ストアに登録中...")
				result, err := trust.Install(s, trust.Options{})
				if err != nil {
					return err
				}
				printTrustResult(result)
				if err := trustExitError(result); err != nil {
					return err
				}
			} else {
				fmt.Println("\n次のステップ: ore2ca trust  # OSに信頼登録")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&certPath, "cert", "", "インポートするCA証明書のパス（必須）")
	cmd.Flags().BoolVar(&force, "force", false, "既存のCAを上書き")
	cmd.Flags().BoolVar(&doTrust, "trust", false, "インポート後に信頼登録まで行う")
	cmd.MarkFlagRequired("cert")

	return cmd
}
