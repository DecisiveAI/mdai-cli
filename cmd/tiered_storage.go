package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

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
	Key             string   `json:"-"`
	Tier            string   `json:"tier"`
	Capacity        string   `json:"capacity"`
	RetentionPeriod string   `json:"retention_period"`
	Format          string   `json:"format"`
	Description     string   `json:"description"`
	Pipelines       []string `json:"pipelines"`
	Location        string   `json:"location"`
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
				f := TieredStorageOutputAddFlags{}
				if err := json.Unmarshal([]byte(t), &f); err != nil {
					return fmt.Errorf("error unmarshaling tiered storage: %w", err)
				}
				row := []string{
					k,
					f.Tier,
					"Enabled",
					f.Description,
					strings.Join(f.Pipelines, ", "),
					f.Capacity,
					f.RetentionPeriod,
					f.Format,
					f.Location,
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
	_, _ = fmt.Fprintf(&sb, `tiered storage added successfully, "%s"`, f.Key)
	fmt.Printf("Key: %s\nTier: %s\nCapacity: %s\nRetention Period: %s\nFormat: %s\nDescription: %s\nPipelines: %v\nLocation: %s\n",
		f.Key, f.Tier, f.Capacity, f.RetentionPeriod, f.Format, f.Description, f.Pipelines, f.Location)
	return sb.String()
}

func NewTieredStorageAddCommand() *cobra.Command {
	f := TieredStorageOutputAddFlags{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a tiered storage",
		Long:  `add a tiered storage`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Name as key for storage tier").
						Value(&f.Key).
						Placeholder("log_cold_storage").
						Validate(func(str string) error {
							if str == "" {
								return errors.New("key cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Tier of storage").
						Value(&f.Tier).
						Placeholder("hot, cold, or glacial").
						Validate(func(str string) error {
							if str == "" {
								return errors.New("tier cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Capacity of storage tier").
						Value(&f.Capacity).
						Placeholder("1000gb").
						Validate(func(str string) error {
							if str == "" {
								return errors.New("capacity cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Retention period of storage tier").
						Value(&f.RetentionPeriod).
						Placeholder("30days").
						Validate(func(str string) error {
							if str == "" {
								return errors.New("retention period cannot be empty")
							}
							return nil
						}),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Format for storage tier").
						Value(&f.Format).
						Placeholder("iceberg").
						Validate(func(str string) error {
							if str == "" {
								return errors.New("format cannot be empty")
							}
							return nil
						}),
					huh.NewInput().
						Title("Location of storage tier").
						Value(&f.Location).
						Validate(func(str string) error {
							if str == "" {
								return errors.New("location cannot be empty")
							}
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
						Value(&f.Pipelines),
				),

				huh.NewGroup(
					huh.NewInput().
						Title("Description (optional)").
						Value(&f.Description).
						Placeholder("This is tiered storage location that will go to S3").
						Validate(func(str string) error {
							return nil
						}),
				),
			)

			err := form.Run()
			if err != nil {
				fmt.Println("Storage tier failed due to", err)
			} else {

				ctx := cmd.Context()
				helper, err := kubehelper.New(kubehelper.WithContext(ctx))
				if err != nil {
					return fmt.Errorf("failed to create kube helper: %w", err)
				}

				jsonData, _ := json.Marshal(f)
				if _, err := helper.UpdateConfigMap(cmd.Context(),
					"tiered-storage", "mdai",
					map[string]string{f.Key: string(jsonData)}); err != nil {
					return fmt.Errorf("failed to add storage tier: %w", err)
				}

				fmt.Println(f.successString())
			}

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
