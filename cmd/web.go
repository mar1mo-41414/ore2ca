package cmd

import (
	"fmt"

	"github.com/mar1mo-41414/ore2ca/internal/web"
	"github.com/spf13/cobra"
)

func newWebCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Web UIを起動する",
		Long:  `ブラウザで証明書の管理・発行・ダウンロードができるローカルWeb UIを起動します。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			srv, err := web.NewServer(addr)
			if err != nil {
				return err
			}
			return srv.Start()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "待受ポート番号")
	return cmd
}
