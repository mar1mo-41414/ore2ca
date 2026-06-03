package cmd

import (
	"fmt"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/mar1mo-41414/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newTrustCmd() *cobra.Command {
	var onlyFirefox, onlyChrome bool

	cmd := &cobra.Command{
		Use:   "trust",
		Short: "CAをOSの信頼ストアに登録する",
		Long: `作成済みのローカルCAをOSのルート証明書ストアに登録します。

ブラウザ固有の登録先（Firefox NSS、Chrome NSS）はインストール済みのものを自動検出します。
--firefox / --chrome を指定した場合は、そのブラウザのみを登録対象にします（OS登録は常に実施）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません。先に ore2ca init を実行してください")
			}

			opts := trust.Options{
				Firefox: onlyFirefox,
				Chrome:  onlyChrome,
			}

			fmt.Println("CAをシステムの信頼ストアに登録中...")
			result, err := trust.Install(s, opts)
			if err != nil {
				return err
			}

			printTrustResult(result)
			return trustExitError(result)
		},
	}

	cmd.Flags().BoolVar(&onlyFirefox, "firefox", false, "Firefox のみブラウザ登録する（OS登録は常に実施）")
	cmd.Flags().BoolVar(&onlyChrome, "chrome", false, "Chrome/Chromium のみブラウザ登録する（OS登録は常に実施）")
	return cmd
}

func printTrustResult(r *trust.Result) {
	if r.OS {
		fmt.Println("✓ OS 信頼ストア: 登録完了")
	} else {
		fmt.Printf("✗ OS 信頼ストア: 失敗 — %v\n", r.OSErr)
	}

	printBrowserResult("Firefox", r.Firefox)
	printBrowserResult("Chrome / Chromium", r.Chrome)

	fmt.Println()
	if r.Firefox.Registered || r.Chrome.Registered {
		fmt.Println("ブラウザを再起動すると変更が反映されます。")
	}
}

func printBrowserResult(name string, br trust.BrowserResult) {
	switch {
	case br.Skipped:
		// 未検出または対象外 — 何も表示しない
	case br.Registered:
		fmt.Printf("✓ %s: 登録完了\n", name)
	default:
		fmt.Printf("✗ %s: 登録失敗\n  %v\n", name, br.Err)
	}
}

func trustExitError(r *trust.Result) error {
	if !r.OS {
		return fmt.Errorf("OS 信頼登録に失敗しました")
	}
	return nil
}
