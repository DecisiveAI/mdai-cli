package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "get -c|--config MODULE-NAME",
		Short:   "get a configuration",
		Long:    "get mdai or otel collector configuration",
		Example: `  mdai get --config mdai # get mdai configuration
  mdai get --config otel # get otel configuration`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			configType, _ := cmd.Flags().GetString("config")
			if configType != "" && !slices.Contains(SupportedGetConfigTypes, configType) {
				return fmt.Errorf("config type %s is not supported", configType)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			configType, _ := cmd.Flags().GetString("config")
			switch configType {
			case "mdai":
				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				fmt.Printf("name           : %s\n", PurpleStyle.Render(get.Name))
				fmt.Printf("namespace      : %s\n", PurpleStyle.Render(get.Namespace))
				fmt.Printf("measure volumes: %v\n", PurpleStyle.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].MeasureVolumes)))
				fmt.Printf("enabled        : %v\n", PurpleStyle.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].Enabled)))
			case "otel":
				get, err := operator.GetOperator(ctx)
				if err != nil {
					return err
				}
				fmt.Println(get.Spec.TelemetryModule.Collectors[0].Spec.Config)
			}

			return nil
		},
	}
	cmd.Flags().StringP("config", "c", "", "configuration to get ["+strings.Join(SupportedGetConfigTypes, ", ")+"]")

	_ = cmd.MarkFlagRequired("config")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
