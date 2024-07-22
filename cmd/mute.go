package cmd

import (
	"fmt"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewMuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "mute -n|--name FILTER-NAME -p|--pipeline PIPELINE-NAME -d|--description DESCRIPTION",
		Short:   "mute a telemetry pipeline",
		Long:    `activate (add to pipeline configuration) a telemetry muting filter`,
		Example: `  mdai mute --name test-filter --description "test filter muting" --pipeline "logs"
  mdai mute --name another-filter --description "metrics pipeline muting" --pipeline "metrics"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filterName, _ := cmd.Flags().GetString("name")
			pipelines, _ := cmd.Flags().GetStringSlice("pipeline")
			description, _ := cmd.Flags().GetString("description")

			if err := operator.Mute(filterName, description, pipelines); err != nil {
				return fmt.Errorf("muting failed: %w", err)
			}
			fmt.Printf("pipeline(s) %v muted successfully as filter %s (%s).\n", pipelines, filterName, description)
			return nil
		},
	}
	cmd.Flags().StringSliceP("pipeline", "p", []string{""}, "pipeline to mute")
	cmd.Flags().StringP("name", "n", "", "name of the filter")
	cmd.Flags().StringP("description", "d", "", "description of the filter")

	cmd.MarkFlagsRequiredTogether("name", "description", "pipeline")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("pipeline")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
