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
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list telemetry filters",
		Long:  `list telemetry filters`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			onlyService, _ := cmd.Flags().GetBool("service")
			onlyPipeline, _ := cmd.Flags().GetBool("pipeline")
			mdaiOperator, err := operator.GetOperator(ctx)
			if err != nil {
				return err
			}

			hasTelemetryFilters := func(mdaiOperator *v1.MyDecisiveEngine) bool {
				return mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering != nil &&
					len(*mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters) > 0
			}

			if !hasTelemetryFilters(mdaiOperator) {
				fmt.Println("No filters found.")
				return nil
			}

			var pipelineFilterRows, filterServicerRows [][]string

			for _, filter := range *mdaiOperator.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
				isServiceFilter := filter.FilteredServices != nil
				if onlyService && filter.FilteredServices == nil {
					continue
				}
				if onlyPipeline && filter.MutedPipelines == nil {
					continue
				}
				var row []string
				row = append(row, filter.Name, filter.Description)
				if filter.Enabled {
					row = append(row, EnabledString)
				} else {
					row = append(row, DisabledString)
				}
				if filter.MutedPipelines != nil {
					row = append(row, strings.Join(*filter.MutedPipelines, ", "))
				}
				if isServiceFilter {
					if filter.FilteredServices.Pipelines != nil {
						row = append(row, strings.Join(*filter.FilteredServices.Pipelines, ", "))
					} else {
						row = append(row, NoDataString)
					}
					if filter.FilteredServices.TelemetryTypes != nil {
						row = append(row, strings.Join(*filter.FilteredServices.TelemetryTypes, ", "))
					} else {
						row = append(row, NoDataString)
					}
					row = append(row, filter.FilteredServices.ServiceNamePattern)
					filterServicerRows = append(filterServicerRows, row)
				} else {
					pipelineFilterRows = append(pipelineFilterRows, row)
				}
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
				fmt.Println(filterTableOutput)
			}

			if !onlyService {
				printFilterTable(pipelineFilterHeaders, pipelineFilterRows)
			}

			if !onlyPipeline {
				printFilterTable(filterServiceHeaders, filterServicerRows)
			}

			return nil
		},
	}

	cmd.Flags().Bool("service", false, "")
	cmd.Flags().Bool("pipeline", false, "")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

var (
	WithName            = operator.WithName
	WithDescription     = operator.WithDescription
	WithService         = operator.WithService
	WithServicePipeline = operator.WithServicePipeline
	WithPipeline        = operator.WithPipeline
	WithTelemetry       = operator.WithTelemetry
)

type filterAddFlags struct {
	name        string
	description string
	pipeline    []string
	service     string
	telemetry   []string
}

func (f filterAddFlags) toTelemetryFilterOptions() []operator.TelemetryFilterOption {
	funcs := []operator.TelemetryFilterOption{
		WithName(f.name),
		WithDescription(f.description),
	}
	if f.service != "" {
		funcs = append(funcs, WithService(f.service))
		if len(f.pipeline) > 0 {
			funcs = append(funcs, WithServicePipeline(f.pipeline))
		}
		if len(f.telemetry) > 0 {
			funcs = append(funcs, WithTelemetry(f.telemetry))
		}
	} else if len(f.pipeline) > 0 {
		funcs = append(funcs, WithPipeline(f.pipeline))
	}

	return funcs
}

func (f filterAddFlags) successString() string {
	var sb strings.Builder
	if f.service != "" {
		_, _ = fmt.Fprintf(&sb, `service pattern "%s" added successfully as filter "%s" (%s)`, f.service, f.name, f.description)
		if len(f.pipeline) > 0 {
			_, _ = fmt.Fprintf(&sb, " for pipelines %v\n", f.pipeline)
		}
		if len(f.telemetry) > 0 {
			_, _ = fmt.Fprintf(&sb, " for telemetry types %v\n", f.telemetry)
		}
	} else {
		_, _ = fmt.Fprintf(&sb, `pipeline(s) %v added successfully as filter "%s" (%s).`, f.pipeline, f.name, f.description)
		_, _ = fmt.Fprintln(&sb)
	}
	return sb.String()
}

func NewFilterAddCommand() *cobra.Command {
	f := filterAddFlags{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a telemetry filter",
		Long:  `add a telemetry filter`,
		Example: `  add --name filter-1 --description filter-1 --pipeline logs
  add --name filter-1 --description filter-1 --pipeline logs --service service-1
  add --name filter-1 --description filter-1 --telemetry logs --service service-1`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if f.service == "" {
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

			if err := operator.CreateTelemetryFilter(ctx, f.toTelemetryFilterOptions()...); err != nil {
				return fmt.Errorf("adding filter failed: %w", err)
			}

			fmt.Println(f.successString())

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&f.pipeline, "pipeline", "p", []string{}, "pipeline to mute")
	cmd.Flags().StringVarP(&f.name, "name", "n", "", "name of the filter")
	cmd.Flags().StringVarP(&f.description, "description", "d", "", "description of the filter")
	cmd.Flags().StringVarP(&f.service, "service", "s", "", "service pattern")
	cmd.Flags().StringSliceVarP(&f.telemetry, "telemetry", "t", []string{}, "telemetry type")

	cmd.MarkFlagsMutuallyExclusive("pipeline", "telemetry")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable",
		Short:   "disable a telemetry filter",
		Long:    `disable a telemetry filter`,
		Example: `  disable --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			filterName, _ := cmd.Flags().GetString("name")

			if err := operator.DisableTelemetryFilter(ctx, WithName(filterName)); err != nil {
				return fmt.Errorf("disabling filter failed: %w", err)
			}
			fmt.Printf(`"%s" filter disabled successfully.`, filterName)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enable",
		Short:   "enable a telemetry filter",
		Long:    `enable a telemetry filter`,
		Example: `  enable --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			filterName, _ := cmd.Flags().GetString("name")

			if err := operator.EnableTelemetryFilter(ctx, WithName(filterName)); err != nil {
				return fmt.Errorf("enabling filter failed: %w", err)
			}
			fmt.Printf(`"%s" filter enabled successfully.`, filterName)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewFilterRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove a telemetry filter",
		Long:    `remove a telemetry filter`,
		Example: `  remove --name filter-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			filterName, _ := cmd.Flags().GetString("name")

			if err := operator.RemoveTelemetryFilter(ctx, WithName(filterName)); err != nil {
				return fmt.Errorf("removing filter failed: %w", err)
			}
			fmt.Printf(`"%s" filter removed successfully.`, filterName)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringP("name", "n", "", "name of the filter")

	_ = cmd.MarkFlagRequired("name")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
