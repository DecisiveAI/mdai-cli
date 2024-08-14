package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	mdaihelm "github.com/decisiveai/mdai-cli/internal/helm"
	mdaitypes "github.com/decisiveai/mdai-cli/internal/types"
	"github.com/spf13/cobra"
)

func NewOutdatedCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "outdated",
		Short:   "shows current and wanted versions of MDAI installation packages",
		Long:    `shows current and wanted versions of MDAI installation packages`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			helmclient := mdaihelm.NewClient(mdaihelm.WithContext(ctx))
			rows, err := helmclient.Outdated()
			if err != nil {
				return fmt.Errorf("failed to fetch package information: %w", err)
			}

			t := table.New().
				BorderHeader(false).
				Border(lipgloss.HiddenBorder()).
				StyleFunc(func(row, col int) lipgloss.Style {
					switch {
					case row == 0:
						return HeaderStyle
					case rows[row-1][col] == "✗":
						return OutdatedStyle
					case rows[row-1][col] == "✓":
						return UpToDateStyle
					case row%2 == 0:
						return EvenRowStyle
					default:
						return OddRowStyle
					}
				}).
				Headers("", "RELEASE", "CURRENT", "WANTED").
				Rows(rows...)
			fmt.Println(t)
			fmt.Printf("kubeconfig: %s\nkubecontext: %s\n",
				PurpleStyle.Render(ctx.Value(mdaitypes.Kubeconfig{}).(string)),
				PurpleStyle.Render(ctx.Value(mdaitypes.Kubecontext{}).(string)),
			)
			return nil
		},
	}
	return cmd
}
