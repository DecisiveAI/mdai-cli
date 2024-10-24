package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/decisiveai/mdai-cli/internal/forms"
	"github.com/decisiveai/mdai-cli/internal/kubehelper"
	"github.com/decisiveai/mdai-cli/internal/types"
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

			headers := []string{"Name", "Tier", "Status", "Description", "Pipelines", "Capacity", "Duration", "Format", "Location"}
			var rows [][]string

			configMap, err := helper.GetConfigMap(ctx, "tiered-storage", "mdai")
			if err != nil {
				return fmt.Errorf("failed to fetch tiered storages configmap: %w", err)
			}
			for k, t := range configMap.Data {
				f := types.TieredStorageOutputAddFlags{}
				if err := json.Unmarshal([]byte(t), &f); err != nil {
					return fmt.Errorf("error unmarshaling tiered storage: %w", err)
				}
				row := []string{
					k,
					f.Tier,
					"Enabled",
					f.Description,
					strings.Join(f.Pipelines, ", "),
					f.Capacity + " " + f.CapacityType,
					f.Duration + " " + f.DurationType,
					f.Format,
					f.Store,
				}
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
						if row == table.HeaderRow {
							return HeaderStyle
						}
						if row%2 == 0 {
							return OddRowStyle
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

func NewTieredStorageAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a tiered storage",
		Long:  `add a tiered storage`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			completed, f := forms.TieredStorageForm()
			if !completed {
				return errors.New("input cancelled")
			}
			ctx := cmd.Context()
			helper, err := kubehelper.New(kubehelper.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to create kube helper: %w", err)
			}

			jsonData, _ := json.Marshal(f)
			if _, err := helper.UpdateConfigMap(cmd.Context(),
				"tiered-storage", "mdai",
				map[string]string{f.Name: string(jsonData)}); err != nil {
				return fmt.Errorf("failed to add storage tier: %w", err)
			}

			fmt.Println(f.SuccessString())
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
