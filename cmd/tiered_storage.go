package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	"github.com/spf13/cobra"
)

// Example Table Output:
/*
+------------+--------+----------+-----------------+----------+-------------------------+------------+-------------+
|    Key     |  Tier  | Capacity | Retention Period |  Format  |      Description         | Pipelines  |  Location   |
+------------+--------+----------+-----------------+----------+-------------------------+------------+-------------+
|   tier_1   |  hot   |   500GB  |     30 days      | iceberg  | Main storage for hot data| logs, traces | some/location        |
|   tier_2   |  cold  |   1TB    |     90 days      | parquet  | Backup cold storage      | logs         | /another/location    |
|   tier_3   | glacial|   5TB    |    365 days      | ORC      | Archival storage         | metrics      | random/fileMCfileFace|
*/

type TieredStorageOutputAddFlags struct {
	key             string
	tier            string
	capacity        string
	retentionPeriod string
	format          string
	description     string
	pipelines       []string
	location        string
}

type TieredStorageValues struct {
}

func NewTieredStorageCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "configuration",
		Use:     "tiered_storage",
		Short:   "manage tiered storages",
		Long:    `manage tiered storages`,
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	cmd.AddCommand(
		NewTieredStorageAddCommand(),
		NewTieredStorageListCommand(),
		NewTieredStorageRemoveCommand(),
	)

	return cmd
}

func NewTieredStorageListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list tiered storages",
		Long:  `list tiered storages`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to initialize kubehelper: %w", err)
			}

			headers := []string{"Key", "Tier", "Status", "Description", "Pipelines", "Capacity", "Retention Period", "Format", "Location"}
			var rows [][]string

			configMap, err := helper.GetConfigMap(ctx, "tiered-storage", "mdai")
			if err != nil {
				return fmt.Errorf("failed to fetch tiered storages configmap: %w", err)
			}
			for k, t := range configMap.Data {
				row := []string{k, t, "Enabled"}
				rows = append(rows, row)
			}

			printTieredStorage := func(headers []string, rows [][]string) {
				if len(rows) == 0 {
					fmt.Println("No tiered storage found.")
					return
				}

				TieredStorageOutput := table.New().
					BorderHeader(false).
					Border(lipgloss.HiddenBorder()).
					StyleFunc(func(row, col int) lipgloss.Style {
						if row == 0 {
							return HeaderStyle
						}
						return lipgloss.NewStyle()
					}).
					Headers(headers...).
					Rows(rows...)
				fmt.Println(TieredStorageOutput)
			}
			printTieredStorage(headers, rows)
			return nil
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func (f TieredStorageOutputAddFlags) successString() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `tiered storage added successfully, "%s"`, f.key)
	fmt.Printf("Key: %s\nTier: %s\nCapacity: %s\nRetention Period: %s\nFormat: %s\nDescription: %s\nPipelines: %v\nLocation: %s\n",
		f.key, f.tier, f.capacity, f.retentionPeriod, f.format, f.description, f.pipelines, f.location)
	return sb.String()
}

func NewTieredStorageAddCommand() *cobra.Command {
	f := TieredStorageOutputAddFlags{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a tiered storage",
		Long:  `add a tiered storage`,

		/*PreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.MarkFlagsRequiredTogether("key", "tier", "capacity", "retention-period", "format", "pipelines", "location")
			return nil
		},*/
		RunE: func(cmd *cobra.Command, _ []string) error {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Name as key for storage tier (ex. some_tier)").
						Value(&f.key).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Key cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Tier of storage").
						Value(&f.tier).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Tier cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Capacity of storage tier").
						Value(&f.capacity).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Capacity cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Retention period of storage tier").
						Value(&f.retentionPeriod).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Retention period cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Format for storage tier (ex. iceberg)").
						Value(&f.format).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Format cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewText().
						Title("Description (optional)").
						Value(&f.description).
						Validate(func(str string) error {
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Pipelines").
						Options(
							huh.NewOption("traces", "traces"),
							huh.NewOption("metrics", "metrics"),
							huh.NewOption("Logs", "logs").Selected(true),
						).
						Limit(3).
						Value(&f.pipelines),
					huh.NewInput().
						Title("Location of storage tier").
						Value(&f.location).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("Location cannot be empty")
							}
							return nil
						}),
				),
			)

			err := form.Run()
			if err != nil {
				fmt.Println("Storage tier failed due to", err)
			}

			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to create kube helper: %w", err)
			}

			if _, err := helper.UpdateConfigMap(cmd.Context(), "tiered-storage", "mdai", map[string]string{"key": f.key, "tier": f.tier, "capacity": f.capacity, "retention-period": f.retentionPeriod, "format": f.format, "description": f.description, "pipelines": strings.Join(f.pipelines, ", "), "location": f.location}); err != nil {
				return fmt.Errorf("failed to add storage tier: %w", err)
			}

			fmt.Println(f.successString())
			fmt.Printf("Key: %s\nTier: %s\nCapacity: %s\nRetention Period: %s\nFormat: %s\nDescription: %s\nPipelines: %v\nLocation: %s\n",
				f.key, f.tier, f.capacity, f.retentionPeriod, f.format, f.description, f.pipelines, f.location)

			return nil
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}

func NewTieredStorageRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove a tiered storage",
		Long:    `remove a tiered storage`,
		Example: `  remove --key some_key`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			TieredStorageOutputKey, _ := cmd.Flags().GetString("key")
			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to initialize kubehelper: %w", err)
			}
			configMap, err := helper.GetConfigMap(ctx, "tiered-storage", "mdai")
			if err != nil {
				return fmt.Errorf("failed to fetch tiered storages configmap: %w", err)
			}
			if _, found := configMap.Data[TieredStorageOutputKey]; !found {
				return fmt.Errorf("tiered storage %s not found in configmap", TieredStorageOutputKey)
			}
			delete(configMap.Data, TieredStorageOutputKey)
			if _, err = helper.SetConfigMap(ctx, "mdai", configMap); err != nil {
				return fmt.Errorf("failed to remove tiered storage: %w", err)
			}

			fmt.Printf(`"%s"tiered storage removed successfully.`, TieredStorageOutputKey)
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringP("key", "k", "", "name of the tiered storage")
	_ = cmd.MarkFlagRequired("key")

	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true

	return cmd
}
