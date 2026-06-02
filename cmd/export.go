package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var outPath string
	var password string
	var legacy bool

	cmd := &cobra.Command{
		Use:   "export <id>",
		Short: "証明書を PKCS#12 (.pfx) 形式でエクスポートする",
		Long: `指定したIDの証明書・秘密鍵・CAチェーンをまとめて .pfx ファイルに書き出します。

Windows (IIS, mmc) や Java アプリへの証明書インポートに使用できます。

例:
  ore2ca export 0001
  ore2ca export 0001 --out myapp.pfx
  ore2ca export 0001 --password secretpass
  ore2ca export 0001 --legacy   # 古いツール向け（3DES/RC2 暗号化）`,
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

			// 証明書を読み込む
			certPEM, err := os.ReadFile(s.CertPath(meta.Domain))
			if err != nil {
				return fmt.Errorf("証明書ファイルの読み込み失敗: %w", err)
			}
			certBlock, _ := pem.Decode(certPEM)
			if certBlock == nil {
				return fmt.Errorf("証明書の PEM デコード失敗")
			}
			cert, err := x509.ParseCertificate(certBlock.Bytes)
			if err != nil {
				return fmt.Errorf("証明書の解析失敗: %w", err)
			}

			// 秘密鍵を読み込む
			keyPEM, err := os.ReadFile(s.KeyPath(meta.Domain))
			if err != nil {
				return fmt.Errorf("秘密鍵ファイルの読み込み失敗: %w", err)
			}
			keyBlock, _ := pem.Decode(keyPEM)
			if keyBlock == nil {
				return fmt.Errorf("秘密鍵の PEM デコード失敗")
			}
			privateKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return fmt.Errorf("秘密鍵の解析失敗: %w", err)
			}

			// CA 証明書を読み込む
			caCertPEM, err := os.ReadFile(s.CACertPath())
			if err != nil {
				return fmt.Errorf("CA証明書ファイルの読み込み失敗: %w", err)
			}
			caBlock, _ := pem.Decode(caCertPEM)
			if caBlock == nil {
				return fmt.Errorf("CA証明書の PEM デコード失敗")
			}
			caCert, err := x509.ParseCertificate(caBlock.Bytes)
			if err != nil {
				return fmt.Errorf("CA証明書の解析失敗: %w", err)
			}

			// PKCS#12 へエンコード
			var encoder *pkcs12.Encoder
			if legacy {
				encoder = pkcs12.LegacyDES // 古いツール（Java 8以前, 旧 Windows 等）向け
			} else {
				encoder = pkcs12.Modern2023 // AES-256 + PBKDF2-SHA256
			}

			pfxData, err := encoder.Encode(privateKey, cert, []*x509.Certificate{caCert}, password)
			if err != nil {
				return fmt.Errorf("PKCS#12 エンコード失敗: %w", err)
			}

			// 出力パスを決定
			if outPath == "" {
				safeDomain := strings.ReplaceAll(meta.Domain, "*", "_wildcard_")
				outPath = filepath.Join(".", safeDomain+".pfx")
			}
			if err := os.WriteFile(outPath, pfxData, 0600); err != nil {
				return fmt.Errorf("ファイル書き込み失敗: %w", err)
			}

			fmt.Printf("✓ エクスポート完了: %s\n", outPath)
			fmt.Printf("  ドメイン:     %s\n", meta.Domain)
			if legacy {
				fmt.Printf("  暗号化形式:   Legacy (3DES) ※古いツール向け\n")
			} else {
				fmt.Printf("  暗号化形式:   Modern 2023 (AES-256 + PBKDF2-SHA256)\n")
			}
			if password == "" {
				fmt.Printf("  パスワード:   なし（空文字列）\n")
			} else {
				fmt.Printf("  パスワード:   設定済み\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outPath, "out", "o", "", "出力ファイルパス（省略時: <domain>.pfx）")
	cmd.Flags().StringVarP(&password, "password", "p", "", "PFX ファイルのパスワード（省略時: 空）")
	cmd.Flags().BoolVar(&legacy, "legacy", false, "古いツール向けの暗号化（3DES）を使用")
	return cmd
}
