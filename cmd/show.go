package cmd

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "証明書の詳細情報を表示する",
		Long:  `指定したIDの証明書の詳細情報（SAN・有効期間・アルゴリズム等）を表示します。`,
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

			certPEM, err := os.ReadFile(s.CertPath(meta.Domain))
			if err != nil {
				return fmt.Errorf("証明書ファイルの読み込み失敗: %w", err)
			}
			block, _ := pem.Decode(certPEM)
			if block == nil {
				return fmt.Errorf("PEMデコード失敗: %s", s.CertPath(meta.Domain))
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("証明書解析失敗: %w", err)
			}

			now := time.Now()
			status := "有効"
			if meta.Revoked {
				status = fmt.Sprintf("失効（%s）", meta.RevokedAt.Format("2006-01-02"))
			} else if now.After(cert.NotAfter) {
				status = "期限切れ"
			} else if now.After(cert.NotAfter.AddDate(0, 0, -30)) {
				status = "期限切れ間近"
			}

			fmt.Println("─────────────────────────────────────────")
			fmt.Printf("  ID:             %s\n", meta.ID)
			fmt.Printf("  ステータス:     %s\n", status)
			fmt.Println("─────────────────────────────────────────")
			fmt.Printf("  コモンネーム:   %s\n", cert.Subject.CommonName)
			fmt.Printf("  組織:           %s\n", strings.Join(cert.Subject.Organization, ", "))
			fmt.Printf("  発行者:         %s\n", cert.Issuer.CommonName)
			fmt.Printf("  シリアル番号:   %s\n", cert.SerialNumber)
			fmt.Println()
			fmt.Printf("  有効期間開始:   %s\n", cert.NotBefore.Local().Format("2006-01-02 15:04:05"))
			fmt.Printf("  有効期間終了:   %s\n", cert.NotAfter.Local().Format("2006-01-02 15:04:05"))
			remaining := time.Until(cert.NotAfter)
			if remaining > 0 {
				fmt.Printf("  残り日数:       %d 日\n", int(remaining.Hours()/24))
			}
			fmt.Println()

			if len(cert.DNSNames) > 0 {
				fmt.Printf("  DNS SAN:        %s\n", strings.Join(cert.DNSNames, ", "))
			}
			if len(cert.IPAddresses) > 0 {
				ips := make([]string, len(cert.IPAddresses))
				for i, ip := range cert.IPAddresses {
					ips[i] = ip.String()
				}
				fmt.Printf("  IP SAN:         %s\n", strings.Join(ips, ", "))
			}
			fmt.Println()

			fmt.Printf("  署名アルゴリズム: %s\n", cert.SignatureAlgorithm)
			fmt.Printf("  公開鍵アルゴリズム: %s", cert.PublicKeyAlgorithm)
			switch pub := cert.PublicKey.(type) {
			case *ecdsa.PublicKey:
				fmt.Printf(" (%s)\n", pub.Curve.Params().Name)
			case *rsa.PublicKey:
				fmt.Printf(" (%d bit)\n", pub.Size()*8)
			default:
				fmt.Println()
			}
			fmt.Println()

			fmt.Printf("  証明書:         %s\n", s.CertPath(meta.Domain))
			fmt.Printf("  秘密鍵:         %s\n", s.KeyPath(meta.Domain))
			fmt.Printf("  フルチェーン:   %s\n", s.FullchainPath(meta.Domain))
			fmt.Println("─────────────────────────────────────────")
			return nil
		},
	}
	return cmd
}
