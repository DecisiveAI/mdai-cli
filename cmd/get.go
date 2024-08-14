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
				fmt.Printf("name           : %s\n", purple.Render(get.Name))
				fmt.Printf("namespace      : %s\n", purple.Render(get.Namespace))
				fmt.Printf("measure volumes: %v\n", purple.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].MeasureVolumes)))
				fmt.Printf("enabled        : %v\n", purple.Render(strconv.FormatBool(get.Spec.TelemetryModule.Collectors[0].Enabled)))
				if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering != nil {
					fmt.Println("filters")
					for _, filter := range *get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
						fmt.Printf("\tname       : %s\n", purple.Render(filter.Name))
						fmt.Printf("\tdescription: %s\n", purple.Render(filter.Description))
						fmt.Printf("\tenabled    : %v\n", purple.Render(strconv.FormatBool(filter.Enabled)))
						fmt.Printf("\tpipelines  : %s\n", purple.Render(strings.Join(*filter.MutedPipelines, ", ")))
						fmt.Println("\t--")
					}
				}
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
