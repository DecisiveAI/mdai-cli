package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/decisiveai/mdai-cli/internal/operator"
	v1 "github.com/decisiveai/mydecisive-engine-operator/api/v1"
	"github.com/spf13/cobra"
)

var (
	WithName            = operator.WithName
	WithDescription     = operator.WithDescription
	WithService         = operator.WithService
	WithServicePipeline = operator.WithServicePipeline
	WithPipeline        = operator.WithPipeline
	WithTelemetry       = operator.WithTelemetry
)

func NewFilterCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "filter",
		Short:   "telemetry filtering",
		Long:    `telemetry filtering`,
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	cmd.AddCommand(
		NewFilterAddCommand(),
		NewFilterDisableCommand(),
		NewFilterEnableCommand(),
		NewFilterListCommand(),
		NewFilterRemoveCommand(),
	)

	return cmd
}

func NewFilterListCommand() *cobra.Command {
	flags := filterListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list telemetry filters",
		Long:  `list telemetry filters`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			mdaiOperator, err := operator.GetOperator(ctx)
			if err != nil {
				return err
			}

			hasTelemetryFilters := func(mdaiOperator *v1.MyDecisiveEngine) bool {
				return mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering != nil &&
					len(*mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters) > 0
			}

			if !hasTelemetryFilters(mdaiOperator) {
				fmt.Fprintln(cmd.OutOrStdout(), "No filters found.")
				return nil
			}

			isFilterEnabled := func(enabled bool) string {
				if enabled {
					return EnabledString
				}
				return DisabledString
			}

			joinOrNoData := func(list *[]string) string {
				if list != nil {
					return strings.Join(*list, ", ")
				}
				return NoDataString
			}

			printFilterTable := func(headers []string, rows [][]string) {
				if len(rows) == 0 {
					return
				}
				filterTableOutput := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						switch {
						case row == 0:
							return HeaderStyle
						case rows[row-1][col] == DisabledString:
							return DisabledStyle.Align(lipgloss.Center)
						case rows[row-1][col] == EnabledString:
							return EnabledStyle.Align(lipgloss.Center)
						case row%2 == 0:
							return EvenRowStyle
						default:
							return OddRowStyle
						}
					}).
					Headers(headers...).
					Rows(rows...)
				fmt.Fprintln(cmd.OutOrStdout(), filterTableOutput)
			}

			var pipelineFilterRows, filterServiceRows [][]string

			for _, filter := range *mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
				if flags.onlyService && filter.FilteredServices == nil {
					continue
				}
				if flags.onlyPipeline && filter.MutedPipelines == nil {
					continue
				}
				var row []string
				row = append(row,
					filter.Name,
					filter.Description,
					isFilterEnabled(filter.Enabled),
				)
				if filter.MutedPipelines != nil {
					row = append(row, strings.Join(*filter.MutedPipelines, ", "))
				}
				switch filter.FilteredServices != nil {
				case true:
					row = append(row,
						joinOrNoData(filter.FilteredServices.Pipelines),
						joinOrNoData(filter.FilteredServices.TelemetryTypes),
						filter.FilteredServices.ServiceNamePattern,
					)
					filterServiceRows = append(filterServiceRows, row)
				case false:
					pipelineFilterRows = append(pipelineFilterRows, row)
				}
			}

			if !flags.onlyService {
				printFilterTable(pipelineFilterHeaders(), pipelineFilterRows)
			}

			if !flags.onlyPipeline {
				printFilterTable(filterServiceHeaders(), filterServiceRows)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.onlyService, "service", false, "")
	cmd.Flags().BoolVar(&flags.onlyPipeline, "pipeline", false, "")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterAddCommand() *cobra.Command {
	flags := filterAddFlags{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a telemetry filter",
		Long:  `add a telemetry filter`,
		Example: `  add --name filter-1 --description filter-1 --pipeline logs
  add --name filter-1 --description filter-1 --pipeline logs --service service-1
  add --name filter-1 --description filter-1 --telemetry logs --service service-1`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if flags.service == "" {
				cmd.MarkFlagsRequiredTogether("name", "description", "pipeline")
			} else {
				cmd.MarkFlagsRequiredTogether("name", "description", "service")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().NFlag() == 0 {
				return errors.New(cmd.UsageString())
			}
			ctx := cmd.Context()

			if err := operator.CreateTelemetryFilter(ctx, flags.toTelemetryFilterOptions()...); err != nil {
				return fmt.Errorf("adding filter failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), flags.successString())

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&flags.pipeline, "pipeline", "p", []string{}, "pipeline to mute")
	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "name of the filter")
	cmd.Flags().StringVarP(&flags.description, "description", "d", "", "description of the filter")
	cmd.Flags().StringVarP(&flags.service, "service", "s", "", "service pattern")
	cmd.Flags().StringSliceVarP(&flags.telemetry, "telemetry", "t", []string{}, "telemetry type")

	cmd.MarkFlagsMutuallyExclusive("pipeline", "telemetry")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterDisableCommand() *cobra.Command {
	flags := filterDisableFlags{}
	cmd := &cobra.Command{
		Use:     "disable",
		Short:   "disable a telemetry filter",
		Long:    `disable a telemetry filter`,
		Example: `  disable --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			if err := operator.DisableTelemetryFilter(ctx, WithName(flags.name)); err != nil {
				return fmt.Errorf("disabling filter failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), `"%s" filter disabled successfully.`, flags.name)
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterEnableCommand() *cobra.Command {
	flags := filterEnableFlags{}
	cmd := &cobra.Command{
		Use:     "enable",
		Short:   "enable a telemetry filter",
		Long:    `enable a telemetry filter`,
		Example: `  enable --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			if err := operator.EnableTelemetryFilter(ctx, WithName(flags.name)); err != nil {
				return fmt.Errorf("enabling filter failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), `"%s" filter enabled successfully.`, flags.name)
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterRemoveCommand() *cobra.Command {
	flags := filterRemoveFlags{}
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove a telemetry filter",
		Long:    `remove a telemetry filter`,
		Example: `  remove --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			if err := operator.RemoveTelemetryFilter(ctx, WithName(flags.name)); err != nil {
				return fmt.Errorf("removing filter failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), `"%s" filter removed successfully.`, flags.name)
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
