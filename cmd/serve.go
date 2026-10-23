package cmd

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/mar1mo-41414/ore2ca/internal/mobileconfig"
	"github.com/mar1mo-41414/ore2ca/internal/store"
	qrterminal "github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
	"rsc.io/qr"
)

const serveIndexHTML = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>ore2ca CA Certificate</title>
<style>
  body { font-family: -apple-system, sans-serif; max-width: 480px; margin: 40px auto; padding: 0 20px; }
  h1 { font-size: 1.4em; }
  .card { border: 1px solid #ddd; border-radius: 8px; padding: 16px; margin: 16px 0; }
  a.btn { display: block; background: #0077ff; color: #fff; padding: 12px 16px;
          border-radius: 6px; text-decoration: none; text-align: center; margin: 8px 0; font-size: 1em; }
  .note { font-size: 0.85em; color: #666; margin-top: 8px; }
</style>
</head>
<body>
<h1>CA 証明書の配布</h1>
<p>このデバイスを <strong>{{.CommonName}}</strong> CA に信頼させるには、以下からダウンロードしてインストールしてください。</p>

<div class="card">
  <strong>iOS / Safari</strong>
  <a class="btn" href="/root.mobileconfig">構成プロファイル (.mobileconfig) をインストール</a>
  <p class="note">ダウンロード後、設定 → 一般 → VPN とデバイス管理 でインストールし、
  続けて設定 → 一般 → 情報 → 証明書信頼設定 で CA をオンにしてください。</p>
</div>

<div class="card">
  <strong>Android / その他</strong>
  <a class="btn" href="/root.crt">CA 証明書 (.crt) をダウンロード</a>
  <p class="note">ダウンロード後、システムの証明書管理画面からインポートしてください。</p>
</div>
</body>
</html>
`

func newServeCmd() *cobra.Command {
	var (
		port int
		once bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "モバイル端末向けにCA証明書を配信する",
		Long: `LAN内の端末にCA証明書を配信するための一時HTTPサーバを起動します。

起動後、ターミナルに表示されるQRコードをスマートフォンで読み取ってアクセスしてください。
iOS向けの .mobileconfig プロファイルと、Android/その他向けの .crt の両方を配信します。

⚠ 信頼できないネットワーク上では実行しないでください。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません。まず ore2ca init を実行してください")
			}

			// root.crt を読み込む
			certPEMBytes, err := os.ReadFile(s.CACertPath())
			if err != nil {
				return fmt.Errorf("CA証明書の読み込みに失敗しました: %w", err)
			}
			block, _ := pem.Decode(certPEMBytes)
			if block == nil {
				return fmt.Errorf("CA証明書のPEMデコードに失敗しました")
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("CA証明書のパースに失敗しました: %w", err)
			}
			certDER := block.Bytes
			commonName := cert.Subject.CommonName

			// .mobileconfig を生成
			mcData, err := mobileconfig.Generate(commonName, certDER)
			if err != nil {
				return fmt.Errorf(".mobileconfig の生成に失敗しました: %w", err)
			}

			// LAN IP 検出
			lanIPs := detectLANIPs()
			if len(lanIPs) == 0 {
				fmt.Println("⚠ LAN IPアドレスが見つかりませんでした。ループバック経由のみアクセス可能です。")
				lanIPs = []string{"127.0.0.1"}
			}

			addr := fmt.Sprintf(":%d", port)
			primaryURL := fmt.Sprintf("http://%s:%d/", lanIPs[0], port)

			// --once 対応: 1回ダウンロードされたらシャットダウン
			var downloadCount atomic.Int32
			srv := &http.Server{Addr: addr}

			shutdownAfterDownload := func() {
				if once && downloadCount.Add(1) >= 1 {
					go srv.Shutdown(context.Background()) //nolint:errcheck
				}
			}

			mux := http.NewServeMux()

			// ランディングページ
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				t := serveIndexHTML
				out := replaceCommonName(t, commonName)
				fmt.Fprint(w, out)
			})

			// CA証明書（Android/Chrome向け）
			mux.HandleFunc("/root.crt", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/x-x509-ca-cert")
				w.Header().Set("Content-Disposition", `attachment; filename="root.crt"`)
				w.Write(certPEMBytes) //nolint:errcheck
				shutdownAfterDownload()
			})

			// iOS mobileconfig プロファイル
			mux.HandleFunc("/root.mobileconfig", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/x-apple-aspen-config")
				w.Header().Set("Content-Disposition", `attachment; filename="root.mobileconfig"`)
				w.Write(mcData) //nolint:errcheck
				shutdownAfterDownload()
			})

			srv.Handler = mux

			// 起動情報を表示
			fmt.Println()
			fmt.Printf("🚀 ore2ca serve — CA: %s\n", commonName)
			fmt.Println()
			fmt.Println("⚠ LAN内の端末からアクセス可能な状態です。信頼できないネットワーク上では実行しないでください。")
			fmt.Println()
			fmt.Println("アクセスURL:")
			for _, ip := range lanIPs {
				fmt.Printf("  http://%s:%d/\n", ip, port)
			}
			fmt.Println()
			if once {
				fmt.Println("（--once モード: 1回ダウンロード後に自動終了します）")
				fmt.Println()
			}
			fmt.Println("QRコード（スマートフォンのカメラで読み取ってください）:")
			fmt.Println()
			qrterminal.GenerateHalfBlock(primaryURL, qr.L, os.Stdout)
			fmt.Println()
			fmt.Println("Ctrl+C で終了")
			fmt.Println()

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("サーバ起動に失敗しました: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 9000, "配信ポート番号")
	cmd.Flags().BoolVar(&once, "once", false, "1回ダウンロードされたら自動終了する")
	return cmd
}

// detectLANIPs returns LAN (private) IPv4 addresses of this machine.
func detectLANIPs() []string {
	privateRanges := []net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	var result []string
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		for _, r := range privateRanges {
			if r.Contains(ip4) {
				result = append(result, ip4.String())
				break
			}
		}
	}
	return result
}

// replaceCommonName replaces {{.CommonName}} in the HTML template.
func replaceCommonName(tmpl, cn string) string {
	result := ""
	search := "{{.CommonName}}"
	for i := 0; i < len(tmpl); {
		idx := -1
		for j := i; j <= len(tmpl)-len(search); j++ {
			if tmpl[j:j+len(search)] == search {
				idx = j
				break
			}
		}
		if idx == -1 {
			result += tmpl[i:]
			break
		}
		result += tmpl[i:idx] + cn
		i = idx + len(search)
	}
	return result
}
