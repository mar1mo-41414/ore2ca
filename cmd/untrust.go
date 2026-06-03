package cmd

import (
	"fmt"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/mar1mo-41414/ore2ca/internal/trust"
	"github.com/spf13/cobra"
)

func newUntrustCmd() *cobra.Command {
	var onlyFirefox, onlyChrome bool

	cmd := &cobra.Command{
		Use:   "untrust",
		Short: "CAをOSの信頼ストアから削除する",
		Long: `ore2ca trust で登録したCA証明書をOSの信頼ストアおよびブラウザから削除します。

--firefox / --chrome を指定した場合は、そのブラウザのみを削除対象にします（OS削除は常に実施）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			if !s.CAExists() {
				return fmt.Errorf("CAが存在しません")
			}

			opts := trust.Options{
				Firefox: onlyFirefox,
				Chrome:  onlyChrome,
			}

			fmt.Println("CAをシステムの信頼ストアから削除中...")
			result, err := trust.Uninstall(s, opts)
			if err != nil {
				return err
			}

			printUntrustResult(result)
			return untrustExitError(result)
		},
	}

	cmd.Flags().BoolVar(&onlyFirefox, "firefox", false, "Firefox のみブラウザ削除する（OS削除は常に実施）")
	cmd.Flags().BoolVar(&onlyChrome, "chrome", false, "Chrome/Chromium のみブラウザ削除する（OS削除は常に実施）")
	return cmd
}

func printUntrustResult(r *trust.UninstallResult) {
	if r.OS {
		fmt.Println("✓ OS 信頼ストア: 削除完了")
	} else {
		fmt.Printf("✗ OS 信頼ストア: 失敗 — %v\n", r.OSErr)
	}

	printBrowserUninstallResult("Firefox", r.Firefox)
	printBrowserUninstallResult("Chrome / Chromium", r.Chrome)

	fmt.Println()
	if r.Firefox.Registered || r.Chrome.Registered {
		fmt.Println("ブラウザを再起動すると変更が反映されます。")
	}
}

func printBrowserUninstallResult(name string, br trust.BrowserResult) {
	switch {
	case br.Skipped:
		// 未検出または対象外 — 何も表示しない
	case br.Registered:
		fmt.Printf("✓ %s: 削除完了\n", name)
	default:
		fmt.Printf("✗ %s: 削除失敗\n  %v\n", name, br.Err)
	}
}

func untrustExitError(r *trust.UninstallResult) error {
	if !r.OS {
		return fmt.Errorf("OS 信頼ストアからの削除に失敗しました")
	}
	return nil
}
