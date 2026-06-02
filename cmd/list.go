package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mar1mo-41414/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "発行済み証明書の一覧を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}

			metas, err := s.ListCerts()
			if err != nil {
				return err
			}

			if len(metas) == 0 {
				fmt.Println("証明書がありません。ore2ca issue <domain> で発行してください。")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tドメイン\t発行日\t有効期限\tステータス")
			fmt.Fprintln(w, "---\t--------\t------\t--------\t----------")

			now := time.Now()
			for _, m := range metas {
				if m.Revoked && !showAll {
					continue
				}
				status := "有効"
				if m.Revoked {
					status = "失効"
				} else if now.After(m.ExpiresAt) {
					status = "期限切れ"
				} else if now.After(m.ExpiresAt.AddDate(0, 0, -30)) {
					status = "期限切れ間近"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					m.ID,
					m.Domain,
					m.IssuedAt.Format("2006-01-02"),
					m.ExpiresAt.Format("2006-01-02"),
					status,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "失効済み証明書も表示")
	return cmd
}
