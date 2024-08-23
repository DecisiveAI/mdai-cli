package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/decisiveai/mdai-cli/internal/operator"
	"github.com/spf13/cobra"
)

func NewFilterCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "filter",
		Short:   "telemetry filtering",
		Long:    `telemetry filtering`,
		// Example: ``,
		/*RunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},*/
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
		// Example: `  list`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			onlyService, _ := cmd.Flags().GetBool("service")
			onlyPipeline, _ := cmd.Flags().GetBool("pipeline")
			get, err := operator.GetOperator(ctx)
			if err != nil {
				return err
			}
			var prows [][]string
			var frows [][]string
			if get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering != nil &&
				len(*get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters) > 0 {
				for _, filter := range *get.Spec.TelemetryModule.Collectors[0].TelemetryFiltering.Filters {
					if onlyService && filter.FilteredServices == nil {
						continue
					}
					if onlyPipeline && filter.MutedPipelines == nil {
						continue
					}
					var row []string
					row = append(row, filter.Name, filter.Description)
					if filter.Enabled {
						row = append(row, "✓")
					} else {
						row = append(row, "✗")
					}
					if filter.MutedPipelines != nil {
						row = append(row, strings.Join(*filter.MutedPipelines, ", "))
					}
					if filter.FilteredServices != nil {
						if filter.FilteredServices.Pipelines != nil {
							row = append(row, strings.Join(*filter.FilteredServices.Pipelines, ", "))
						} else {
							row = append(row, "")
						}
						if filter.FilteredServices.TelemetryTypes != nil {
							row = append(row, strings.Join(*filter.FilteredServices.TelemetryTypes, ", "))
						} else {
							row = append(row, "")
						}
						row = append(row, filter.FilteredServices.ServiceNamePattern)
					}
					if len(row) == 4 {
						prows = append(prows, row)
					} else {
						frows = append(frows, row)
					}
				}
			}

			if !onlyService {
				pt := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						switch {
						case row == 0:
							return HeaderStyle
						case prows[row-1][col] == "✗":
							return OutdatedStyle.Align(lipgloss.Center)
						case prows[row-1][col] == "✓":
							return UpToDateStyle.Align(lipgloss.Center)
						case row%2 == 0:
							return EvenRowStyle
						default:
							return OddRowStyle
						}
					}).
					Headers("NAME", "DESCRIPTION", "ENABLED", "MUTED PIPELINES").
					Rows(prows...)
				fmt.Println(pt)
			}

			if !onlyPipeline {
				ft := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						switch {
						case row == 0:
							return HeaderStyle
						case frows[row-1][col] == "✗":
							return OutdatedStyle.Align(lipgloss.Center)
						case frows[row-1][col] == "✓":
							return UpToDateStyle.Align(lipgloss.Center)
						case row%2 == 0:
							return EvenRowStyle
						default:
							return OddRowStyle
						}
					}).
					Headers("NAME", "DESCRIPTION", "ENABLED", "FILTERED PIPELINES", "FILTERED TELEMETRY", "SERVICE PATTERN").
					Rows(frows...)
				fmt.Println(ft)
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

func NewFilterAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add telemetry filter",
		Long:  ``,
		Example: `  add --name filter-1 --description filter-1 --pipeline logs
  add --name filter-1 --description filter-1 --pipeline logs --service service-1
  add --name filter-1 --description filter-1 --telemetry logs --service service-1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			filterName, _ := cmd.Flags().GetString("name")
			pipelines, _ := cmd.Flags().GetStringSlice("pipeline")
			description, _ := cmd.Flags().GetString("description")
			servicePattern, _ := cmd.Flags().GetString("service")
			telemetryTypes, _ := cmd.Flags().GetStringSlice("telemetry")

			funcs := []operator.TelemetryFilterOption{
				WithName(filterName),
				WithDescription(description),
			}
			if servicePattern != "" {
				funcs = append(funcs, WithService(servicePattern))
				if len(pipelines) > 0 {
					funcs = append(funcs, WithServicePipeline(pipelines))
				}
				if len(telemetryTypes) > 0 {
					funcs = append(funcs, WithTelemetry(telemetryTypes))
				}
			} else {
				if len(pipelines) > 0 {
					funcs = append(funcs, WithPipeline(pipelines))
				}
			}

			if err := operator.CreateTelemetryFilter(ctx, funcs...); err != nil {
				return fmt.Errorf("adding filter failed: %w", err)
			}

			if servicePattern != "" {
				fmt.Printf(`service pattern "%s" added successfully as filter "%s" (%s)`, servicePattern, filterName, description)
				if len(pipelines) > 0 {
					fmt.Printf(" for pipelines %v\n", pipelines)
				}
				if len(telemetryTypes) > 0 {
					fmt.Printf(" for telemetry types %v\n", telemetryTypes)
				}
			} else {
				fmt.Printf(`pipeline(s) %v added successfully as filter "%s" (%s).`, pipelines, filterName, description)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringSliceP("pipeline", "p", []string{""}, "pipeline to mute")
	cmd.Flags().StringP("name", "n", "", "name of the filter")
	cmd.Flags().StringP("description", "d", "", "description of the filter")
	cmd.Flags().StringP("service", "s", "", "service pattern")
	cmd.Flags().StringSliceP("telemetry", "t", []string{""}, "telemetry type")

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
