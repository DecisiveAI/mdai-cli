package cmd

import (
	"fmt"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewUnmuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "unmute -n|--name FILTER-NAME",
		Short:   "unmute a telemetry muting filter",
		Long:    `deactivate (delete from pipeline configuration) a telemetry muting filter`,
		Example: `  mdai unmute --name test-filter # unmute the filter with name test-filter`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			filterName, _ := cmd.Flags().GetString("name")
			removeFilter, _ := cmd.Flags().GetBool("remove")
			action := "unmuted"
			if removeFilter {
				action = "removed"
			}

			if err := operator.Unmute(ctx, filterName, removeFilter); err != nil {
				return fmt.Errorf("unmuting failed: %w", err)
			}
			fmt.Printf("%s filter %s successfully.\n", filterName, action)
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "name of the filter")
	cmd.Flags().Bool("remove", false, "remove the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
