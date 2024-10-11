package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewGetCommand() *cobra.Command {
	flags := getFlags{}
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "get -c|--config MODULE-NAME",
		Short:   "get a configuration",
		Long:    "get mdai or otel collector configuration",
		Example: `  mdai get --config mdai # get mdai configuration
  mdai get --config otel # get otel configuration`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			switch flags.config {
			case "mdai":
				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "name           : %s\n", PurpleStyle.Render(get.Name))
				fmt.Fprintf(cmd.OutOrStdout(), "namespace      : %s\n", PurpleStyle.Render(get.Namespace))
				fmt.Fprintf(cmd.OutOrStdout(), "measure volumes: %v\n", PurpleStyle.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].MeasureVolumes)))
				fmt.Fprintf(cmd.OutOrStdout(), "enabled        : %v\n", PurpleStyle.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].Enabled)))
			case "otel":
				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), get.Spec.TelemetryModule.Collectors[0].Spec.Config)
			default:
				return fmt.Errorf("config type %s is not supported", flags.config)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.config, "config", "c", "", "configuration to get ["+strings.Join(supportedGetConfigTypes(), ", ")+"]")

	_ = cmd.MarkFlagRequired("config")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
